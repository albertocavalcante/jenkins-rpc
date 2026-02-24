package io.albertocavalcante.jenkins.steprpc

data class OperationDefinition(
    val name: String,
    val description: String,
    val executionMode: ExecutionMode,
)

enum class ExecutionMode {
    DIRECT,
    CPS_BRIDGE_REQUIRED,
}

class OperationRegistry(private val allowlist: Set<String> = emptySet()) {
    fun isAllowed(operation: String): Boolean {
        if (allowlist.isEmpty()) {
            return true
        }
        return allowlist.contains(operation)
    }

    fun catalog(discovered: List<OperationDefinition>): List<OperationDefinition> {
        if (allowlist.isEmpty()) {
            return discovered.sortedBy { it.name }
        }

        val byName = discovered.associateBy { it.name }
        return allowlist.sorted().map { name ->
            byName[name] ?: OperationDefinition(
                name = name,
                description = "Allowlisted operation not discovered on this controller",
                executionMode = ExecutionMode.CPS_BRIDGE_REQUIRED,
            )
        }
    }
}
