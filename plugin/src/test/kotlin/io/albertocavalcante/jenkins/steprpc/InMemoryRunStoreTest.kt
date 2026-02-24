package io.albertocavalcante.jenkins.steprpc

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class InMemoryRunStoreTest {
    @Test
    fun `create stores retrievable record`() {
        val store = InMemoryRunStore()
        val created = store.create(
            requestId = "req-1",
            runId = "run-1",
            operation = "archiveArtifacts",
            state = "succeeded",
        )

        val loaded = store.get("run-1")
        assertNotNull(loaded)
        assertEquals(created.requestId, loaded.requestId)
        assertEquals("succeeded", loaded.state)
    }

    @Test
    fun `update mutates state and error`() {
        val store = InMemoryRunStore()
        store.create(
            requestId = "req-1",
            runId = "run-1",
            operation = "junit",
            state = "queued",
        )

        val updated = store.update(
            runId = "run-1",
            state = "failed",
            errorCode = "operation_failed",
            errorMessage = "boom",
        )
        assertNotNull(updated)
        assertEquals("failed", updated.state)
        assertEquals("operation_failed", updated.errorCode)
        assertEquals("boom", updated.errorMessage)
        assertNull(store.update(runId = "missing", state = "failed"))
    }

    @Test
    fun `unknown runId returns null`() {
        val store = InMemoryRunStore()
        assertNull(store.get("nonexistent"))
    }

    @Test
    fun `all fields preserved on create`() {
        val store = InMemoryRunStore()
        val created = store.create(
            requestId = "req-1",
            runId = "run-1",
            operation = "archiveArtifacts",
            state = "queued",
            errorCode = "init_error",
            errorMessage = "something failed",
        )

        val loaded = store.get("run-1")
        assertNotNull(loaded)
        assertEquals("req-1", loaded.requestId)
        assertEquals("run-1", loaded.runId)
        assertEquals("archiveArtifacts", loaded.operation)
        assertEquals("queued", loaded.state)
        assertEquals("init_error", loaded.errorCode)
        assertEquals("something failed", loaded.errorMessage)
        assertNotNull(loaded.createdAt)
    }

    @Test
    fun `error fields cleared on update without errors`() {
        val store = InMemoryRunStore()
        store.create(
            requestId = "req-1",
            runId = "run-1",
            operation = "junit",
            state = "queued",
            errorCode = "init_error",
            errorMessage = "old error",
        )

        val updated = store.update(
            runId = "run-1",
            state = "succeeded",
        )
        assertNotNull(updated)
        assertEquals("succeeded", updated.state)
        assertNull(updated.errorCode)
        assertNull(updated.errorMessage)
    }
}
