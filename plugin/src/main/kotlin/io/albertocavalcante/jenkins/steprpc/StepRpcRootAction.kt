package io.albertocavalcante.jenkins.steprpc

import hudson.Extension
import hudson.model.UnprotectedRootAction
import jenkins.model.Jenkins
import org.kohsuke.stapler.HttpResponse

@Extension
class StepRpcRootAction : UnprotectedRootAction {
    private val runStore = InMemoryRunStore()
    private val cpsBridgeQueue = CpsBridgeQueue()
    private val operationRegistry = OperationRegistry()
    private val executor = InRunOperationExecutor(cpsBridgeQueue)
    private val v1Api = StepRpcV1Api(runStore, operationRegistry, executor, cpsBridgeQueue)

    override fun getIconFileName(): String? = null

    override fun getDisplayName(): String = "Step RPC"

    override fun getUrlName(): String = "step-rpc"

    fun doIndex(): HttpResponse {
        Jenkins.get().checkPermission(Jenkins.READ)
        return jsonResponse(
            mapOf(
                "service" to "jenkins-step-rpc-plugin",
                "version" to "0.1.0-SNAPSHOT",
                "status" to "ok",
                "api" to listOf("v1"),
            ),
        )
    }

    fun getV1(): StepRpcV1Api = v1Api
}
