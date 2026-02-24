package io.albertocavalcante.jenkins.steprpc

import hudson.model.User
import jenkins.model.Jenkins
import java.util.logging.Logger

object AuditLogger {
    private val logger: Logger = Logger.getLogger("io.albertocavalcante.jenkins.steprpc.audit")

    private val sensitivePatterns = setOf(
        "password",
        "secret",
        "token",
        "key",
        "credential",
        "api_key",
        "apikey",
        "access_token",
        "private_key",
    )

    fun log(action: String, fields: Map<String, String?> = emptyMap()) {
        val caller = resolveCallerIdentity()
        val parts = mutableListOf("action=$action", "caller=$caller")
        for ((k, v) in fields) {
            parts.add("$k=${v ?: ""}")
        }
        logger.info(parts.joinToString(" "))
    }

    fun redactSensitiveArgs(args: Map<String, Any?>): Map<String, Any?> {
        return args.mapValues { (key, value) -> redactValue(key, value) }
    }

    private fun redactValue(key: String, value: Any?): Any? {
        if (isSensitiveKey(key)) return "***"
        return when (value) {
            is Map<*, *> -> {
                @Suppress("UNCHECKED_CAST")
                val mapValue = value as Map<String, Any?>
                mapValue.mapValues { (k, v) -> redactValue(k, v) }
            }
            is List<*> -> value.map { redactValue(key, it) }
            else -> value
        }
    }

    internal fun isSensitiveKey(key: String): Boolean {
        val lower = key.lowercase()
        return sensitivePatterns.any { lower.contains(it) }
    }

    private fun resolveCallerIdentity(): String {
        val user = User.current()
        if (user != null) return user.id

        return try {
            val auth = Jenkins.getAuthentication2()
            val authName = auth?.name
            if (!authName.isNullOrBlank() && authName != "anonymous") authName else "anonymous"
        } catch (_: Exception) {
            "anonymous"
        }
    }
}
