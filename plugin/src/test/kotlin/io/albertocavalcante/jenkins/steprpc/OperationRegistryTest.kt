package io.albertocavalcante.jenkins.steprpc

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class OperationRegistryTest {
    @Test
    fun `empty allowlist allows any operation`() {
        val registry = OperationRegistry()
        assertTrue(registry.isAllowed("archiveArtifacts"))
        assertTrue(registry.isAllowed("unknownOperation"))
    }

    @Test
    fun `explicit allowlist restricts operations`() {
        val registry = OperationRegistry(setOf("archiveArtifacts"))
        assertTrue(registry.isAllowed("archiveArtifacts"))
        assertFalse(registry.isAllowed("junit"))
    }

    @Test
    fun `catalog applies allowlist policy`() {
        val registry = OperationRegistry(setOf("archiveArtifacts"))
        val catalog = registry.catalog(
            discovered = listOf(
                OperationDefinition("archiveArtifacts", "Archive", ExecutionMode.DIRECT),
                OperationDefinition("junit", "Publish", ExecutionMode.CPS_BRIDGE_REQUIRED),
            ),
        )
        assertEquals(1, catalog.size)
        assertEquals("archiveArtifacts", catalog.first().name)
        assertEquals(ExecutionMode.DIRECT, catalog.first().executionMode)
    }

    @Test
    fun `undiscovered allowlisted operation gets placeholder`() {
        val registry = OperationRegistry(setOf("missingOp"))
        val catalog = registry.catalog(discovered = emptyList())
        assertEquals(1, catalog.size)
        assertEquals("missingOp", catalog.first().name)
        assertEquals(ExecutionMode.CPS_BRIDGE_REQUIRED, catalog.first().executionMode)
        assertTrue(catalog.first().description.contains("not discovered"))
    }

    @Test
    fun `empty allowlist returns all discovered sorted`() {
        val registry = OperationRegistry()
        val catalog = registry.catalog(
            discovered = listOf(
                OperationDefinition("zeta", "Z", ExecutionMode.DIRECT),
                OperationDefinition("alpha", "A", ExecutionMode.DIRECT),
                OperationDefinition("mid", "M", ExecutionMode.DIRECT),
            ),
        )
        assertEquals(listOf("alpha", "mid", "zeta"), catalog.map { it.name })
    }

    @Test
    fun `explicit allowlist rejects unlisted operation`() {
        val registry = OperationRegistry(setOf("archiveArtifacts"))
        assertFalse(registry.isAllowed("deleteWorkspace"))
    }
}
