package io.albertocavalcante.jenkins.steprpc

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class AuditLoggerTest {
    @Test
    fun `sensitive keys are detected`() {
        assertTrue(AuditLogger.isSensitiveKey("password"))
        assertTrue(AuditLogger.isSensitiveKey("userPassword"))
        assertTrue(AuditLogger.isSensitiveKey("SECRET_VALUE"))
        assertTrue(AuditLogger.isSensitiveKey("auth_token"))
        assertTrue(AuditLogger.isSensitiveKey("api_key"))
        assertTrue(AuditLogger.isSensitiveKey("apiKey"))
        assertTrue(AuditLogger.isSensitiveKey("privateKey"))
        assertTrue(AuditLogger.isSensitiveKey("credential"))
        assertTrue(AuditLogger.isSensitiveKey("credentialId"))
        assertTrue(AuditLogger.isSensitiveKey("access_token"))
        assertTrue(AuditLogger.isSensitiveKey("PRIVATE_KEY"))
    }

    @Test
    fun `non-sensitive keys are not flagged`() {
        assertFalse(AuditLogger.isSensitiveKey("name"))
        assertFalse(AuditLogger.isSensitiveKey("operation"))
        assertFalse(AuditLogger.isSensitiveKey("artifacts"))
        assertFalse(AuditLogger.isSensitiveKey("testResults"))
    }

    @Test
    fun `redact top-level sensitive args`() {
        val args = mapOf<String, Any?>(
            "operation" to "build",
            "password" to "hunter2",
            "token" to "abc123",
        )
        val redacted = AuditLogger.redactSensitiveArgs(args)
        assertEquals("build", redacted["operation"])
        assertEquals("***", redacted["password"])
        assertEquals("***", redacted["token"])
    }

    @Test
    fun `redact nested sensitive args`() {
        val args = mapOf<String, Any?>(
            "config" to mapOf<String, Any?>(
                "url" to "https://example.com",
                "secretKey" to "supersecret",
            ),
        )
        val redacted = AuditLogger.redactSensitiveArgs(args)
        @Suppress("UNCHECKED_CAST")
        val nested = redacted["config"] as Map<String, Any?>
        assertEquals("https://example.com", nested["url"])
        assertEquals("***", nested["secretKey"])
    }

    @Test
    fun `redact preserves non-sensitive values`() {
        val args = mapOf<String, Any?>(
            "artifacts" to "build/*.jar",
            "allowEmptyArchive" to true,
            "count" to 42.0,
            "nothing" to null,
        )
        val redacted = AuditLogger.redactSensitiveArgs(args)
        assertEquals("build/*.jar", redacted["artifacts"])
        assertEquals(true, redacted["allowEmptyArchive"])
        assertEquals(42.0, redacted["count"])
        assertEquals(null, redacted["nothing"])
    }

    @Test
    fun `redact handles lists with sensitive parent key`() {
        val args = mapOf<String, Any?>(
            "credentials" to listOf("cred1", "cred2"),
        )
        val redacted = AuditLogger.redactSensitiveArgs(args)
        assertEquals("***", redacted["credentials"])
    }
}
