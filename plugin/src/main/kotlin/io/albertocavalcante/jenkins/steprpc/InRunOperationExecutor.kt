package io.albertocavalcante.jenkins.steprpc

import hudson.AbortException
import hudson.EnvVars
import hudson.FilePath
import hudson.Launcher
import hudson.model.Descriptor
import hudson.model.Job
import hudson.model.Node
import hudson.model.Run
import hudson.model.TaskListener
import hudson.tasks.Builder
import hudson.tasks.Publisher
import java.io.File
import java.lang.reflect.Method
import java.time.Instant
import java.util.UUID
import jenkins.model.Jenkins
import jenkins.tasks.SimpleBuildStep
import net.sf.json.JSONObject
import org.jenkinsci.Symbol

data class InRunExecutionResult(
    val runId: String,
    val state: String,
    val createdAt: Instant,
    val errorCode: String? = null,
    val errorMessage: String? = null,
)

class InRunOperationExecutor(private val cpsBridgeQueue: CpsBridgeQueue) {
    fun execute(requestId: String, operation: String, args: Map<String, Any?>): InRunExecutionResult {
        val runContext = parseRunContext(args)
        val run = resolveRun(runContext)
        val rpcRunId = "rpc-${UUID.randomUUID().toString().substring(0, 12)}"
        val listener = TaskListener.NULL
        val node = resolveNode(runContext)
        val workspace = resolveWorkspace(runContext, node)
        val launcher = node.createLauncher(listener)
        val env = run.getEnvironment(listener)
        val stepArgs = args.filterKeys { it != RUN_CONTEXT_KEY }

        val descriptor = resolveDescriptor(operation)
        if (descriptor == null) {
            val knownPipelineOperation = discoverPipelineStepOperations().containsKey(operation)
            if (knownPipelineOperation) {
                cpsBridgeQueue.enqueue(
                    PendingBridgeRequest(
                        requestId = requestId,
                        runId = rpcRunId,
                        operation = operation,
                        args = stepArgs,
                        targetRunExternalizableId = run.externalizableId,
                    ),
                )
                return InRunExecutionResult(
                    runId = rpcRunId,
                    state = "queued",
                    createdAt = Instant.now(),
                )
            }
            return InRunExecutionResult(
                runId = rpcRunId,
                state = "failed",
                createdAt = Instant.now(),
                errorCode = "operation_not_found",
                errorMessage = "operation '$operation' was not found among installed Jenkins operations",
            )
        }

        return try {
            val step = instantiateSimpleBuildStep(descriptor, stepArgs)
            executeStep(step, run, workspace, env, launcher, listener)
            InRunExecutionResult(
                runId = rpcRunId,
                state = "succeeded",
                createdAt = Instant.now(),
            )
        } catch (e: Exception) {
            InRunExecutionResult(
                runId = rpcRunId,
                state = "failed",
                createdAt = Instant.now(),
                errorCode = "operation_failed",
                errorMessage = e.message ?: "operation execution failed",
            )
        }
    }

    fun discoverOperations(): List<OperationDefinition> {
        val direct = simpleBuildStepDescriptors()
            .flatMap { descriptor ->
                descriptorSymbols(descriptor).map { symbol ->
                    symbol to descriptor.displayName
                }
            }
            .distinctBy { it.first }
            .associate { it.first to OperationDefinition(it.first, "${it.second} (direct)", ExecutionMode.DIRECT) }

        val allPipeline = discoverPipelineStepOperations()
        val cpsOnly = allPipeline
            .filterKeys { !direct.containsKey(it) }
            .mapValues { (name, displayName) ->
                OperationDefinition(name, "$displayName (CPS context required)", ExecutionMode.CPS_BRIDGE_REQUIRED)
            }

        return (direct + cpsOnly).values
            .sortedBy { it.name }
    }

    private fun resolveDescriptor(operation: String): Descriptor<*>? {
        val operations = linkedMapOf<String, Descriptor<*>>()
        simpleBuildStepDescriptors().forEach { descriptor ->
            descriptorOperationNames(descriptor).forEach { name ->
                if (!operations.containsKey(name)) {
                    operations[name] = descriptor
                }
            }
        }
        return operations[operation]
    }

    private fun descriptorOperationNames(descriptor: Descriptor<*>): List<String> {
        val names = linkedSetOf<String>()
        names.addAll(descriptorSymbols(descriptor))
        names.add(descriptor.id)
        names.add(descriptor.clazz.name)
        return names.filter { it.isNotBlank() }
    }

    private fun descriptorSymbols(descriptor: Descriptor<*>): List<String> {
        val symbols = linkedSetOf<String>()
        descriptor.javaClass.getAnnotation(Symbol::class.java)?.value?.let { symbols.addAll(it) }
        descriptor.clazz.getAnnotation(Symbol::class.java)?.value?.let { symbols.addAll(it) }
        return symbols.toList()
    }

    private fun simpleBuildStepDescriptors(): List<Descriptor<*>> {
        val jenkins = Jenkins.get()
        val descriptors = linkedSetOf<Descriptor<*>>()
        jenkins.getDescriptorList(Builder::class.java).forEach { descriptor ->
            if (SimpleBuildStep::class.java.isAssignableFrom(descriptor.clazz)) {
                descriptors.add(descriptor)
            }
        }
        jenkins.getDescriptorList(Publisher::class.java).forEach { descriptor ->
            if (SimpleBuildStep::class.java.isAssignableFrom(descriptor.clazz)) {
                descriptors.add(descriptor)
            }
        }
        return descriptors.toList()
    }

    private fun discoverPipelineStepOperations(): Map<String, String> {
        return try {
            val extensionList = Jenkins.get().getExtensionList("org.jenkinsci.plugins.workflow.steps.StepDescriptor")
            val operations = linkedMapOf<String, String>()
            extensionList.forEach { descriptor ->
                val functionName = invokeNoArgString(descriptor, "getFunctionName") ?: return@forEach
                if (functionName.isBlank()) {
                    return@forEach
                }
                val displayName = invokeNoArgString(descriptor, "getDisplayName")
                    ?: "Pipeline step"
                operations.putIfAbsent(functionName, displayName)
            }
            operations
        } catch (_: Exception) {
            emptyMap()
        }
    }

    private fun invokeNoArgString(instance: Any, methodName: String): String? {
        return try {
            val method = instance.javaClass.getMethod(methodName)
            method.invoke(instance)?.toString()
        } catch (_: Exception) {
            null
        }
    }

    private fun instantiateSimpleBuildStep(descriptor: Descriptor<*>, args: Map<String, Any?>): SimpleBuildStep {
        val instance = instantiateWithDescribableModel(descriptor, args) ?: instantiateWithDescriptor(descriptor, args)
        if (instance !is SimpleBuildStep) {
            throw AbortException("operation '${descriptor.id}' is not a SimpleBuildStep")
        }
        return instance
    }

    private fun instantiateWithDescribableModel(descriptor: Descriptor<*>, args: Map<String, Any?>): Any? {
        return try {
            val modelClass = Class.forName("org.jenkinsci.plugins.structs.describable.DescribableModel")
            val ofMethod = modelClass.getMethod("of", Class::class.java)
            val model = ofMethod.invoke(null, descriptor.clazz)
            val instantiateMethod = modelClass.methods.firstOrNull {
                it.name == "instantiate" && it.parameterTypes.size == 1 && Map::class.java.isAssignableFrom(it.parameterTypes[0])
            } ?: return null
            instantiateMethod.invoke(model, args)
        } catch (_: ClassNotFoundException) {
            null
        }
    }

    private fun instantiateWithDescriptor(descriptor: Descriptor<*>, args: Map<String, Any?>): Any {
        val payload = JSONObject.fromObject(args)
        val method = findDescriptorNewInstanceMethod(descriptor)
            ?: throw AbortException("descriptor '${descriptor.id}' does not expose newInstance(req,json)")
        return method.invoke(descriptor, null, payload)
    }

    private fun findDescriptorNewInstanceMethod(descriptor: Descriptor<*>): Method? {
        return descriptor.javaClass.methods.firstOrNull { method ->
            method.name == "newInstance" &&
                method.parameterTypes.size == 2 &&
                method.parameterTypes[1] == JSONObject::class.java
        }
    }

    private fun executeStep(
        step: SimpleBuildStep,
        run: Run<*, *>,
        workspace: FilePath,
        env: EnvVars,
        launcher: Launcher,
        listener: TaskListener,
    ) {
        if (step.requiresWorkspace()) {
            workspace.mkdirs()
            step.perform(run, workspace, env, launcher, listener)
            return
        }

        step.perform(run, env, listener)
    }

    private fun parseRunContext(args: Map<String, Any?>): RunContext {
        val runContextValue = args[RUN_CONTEXT_KEY]
            ?: throw AbortException("args.$RUN_CONTEXT_KEY is required")
        if (runContextValue !is Map<*, *>) {
            throw AbortException("args.$RUN_CONTEXT_KEY must be an object")
        }

        val runExternalizableID = runContextValue["runExternalizableId"]?.toString()
        val jobFullName = runContextValue["jobFullName"]?.toString()
        val buildNumber = runContextValue["buildNumber"]?.toString()?.toIntOrNull()
        val nodeName = runContextValue["nodeName"]?.toString()
        val workspace = runContextValue["workspace"]?.toString()

        if (!runExternalizableID.isNullOrBlank()) {
            if (nodeName.isNullOrBlank() || workspace.isNullOrBlank()) {
                throw AbortException("args.$RUN_CONTEXT_KEY.nodeName and args.$RUN_CONTEXT_KEY.workspace are required")
            }
            return RunContext(
                runExternalizableID = runExternalizableID,
                jobFullName = null,
                buildNumber = null,
                nodeName = nodeName,
                workspace = workspace,
            )
        }

        if (jobFullName.isNullOrBlank() || buildNumber == null || nodeName.isNullOrBlank() || workspace.isNullOrBlank()) {
            throw AbortException(
                "args.$RUN_CONTEXT_KEY must include either runExternalizableId or jobFullName/buildNumber plus nodeName/workspace",
            )
        }
        return RunContext(
            runExternalizableID = null,
            jobFullName = jobFullName,
            buildNumber = buildNumber,
            nodeName = nodeName,
            workspace = workspace,
        )
    }

    private fun resolveRun(runContext: RunContext): Run<*, *> {
        runContext.runExternalizableID?.let {
            return Run.fromExternalizableId(it)
                ?: throw AbortException("no run found for runExternalizableId '$it'")
        }

        val jenkins = Jenkins.get()
        val job = jenkins.getItemByFullName(runContext.jobFullName!!, Job::class.java)
            ?: throw AbortException("no job found for '${runContext.jobFullName}'")
        return job.getBuildByNumber(runContext.buildNumber!!)
            ?: throw AbortException("no build #${runContext.buildNumber} for '${runContext.jobFullName}'")
    }

    private fun resolveNode(runContext: RunContext): Node {
        val jenkins = Jenkins.get()
        val nodeName = runContext.nodeName!!
        if (nodeName == "built-in" || nodeName == "master" || nodeName == jenkins.selfLabel.name) {
            return jenkins
        }
        return jenkins.getNode(nodeName)
            ?: throw AbortException("no node found with name '$nodeName'")
    }

    private fun resolveWorkspace(runContext: RunContext, node: Node): FilePath {
        val workspace = runContext.workspace!!
        return if (node is Jenkins) {
            FilePath(File(workspace))
        } else {
            val channel = node.channel ?: throw AbortException("node '${node.nodeName}' has no active channel")
            FilePath(channel, workspace)
        }
    }
}

private data class RunContext(
    val runExternalizableID: String?,
    val jobFullName: String?,
    val buildNumber: Int?,
    val nodeName: String?,
    val workspace: String?,
)

private const val RUN_CONTEXT_KEY = "runContext"
