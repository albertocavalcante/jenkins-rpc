package io.albertocavalcante.jenkins.steprpc

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class CpsBridgeQueueTest {
    @Test
    fun `enqueue and complete request lifecycle`() {
        val queue = CpsBridgeQueue()
        val req = PendingBridgeRequest(
            requestId = "req-1",
            runId = "rpc-1",
            operation = "junit",
            args = mapOf("testResults" to "**/*.xml"),
            targetRunExternalizableId = "job/demo#1",
        )

        queue.enqueue(req)
        val pending = queue.nextForRun("job/demo#1")
        assertNotNull(pending)
        assertEquals("rpc-1", pending.runId)

        val completed = queue.complete("rpc-1")
        assertNotNull(completed)
        assertEquals("req-1", completed.requestId)
        assertNull(queue.nextForRun("job/demo#1"))
    }

    @Test
    fun `unknown runId returns null on complete`() {
        val queue = CpsBridgeQueue()
        assertNull(queue.complete("nonexistent"))
    }

    @Test
    fun `empty queue returns null for nextForRun`() {
        val queue = CpsBridgeQueue()
        assertNull(queue.nextForRun("job/demo#1"))
    }

    @Test
    fun `FIFO ordering with multiple requests`() {
        val queue = CpsBridgeQueue()
        val req1 = PendingBridgeRequest(
            requestId = "req-1",
            runId = "rpc-1",
            operation = "junit",
            args = emptyMap(),
            targetRunExternalizableId = "job/demo#1",
        )
        val req2 = PendingBridgeRequest(
            requestId = "req-2",
            runId = "rpc-2",
            operation = "archiveArtifacts",
            args = emptyMap(),
            targetRunExternalizableId = "job/demo#1",
        )

        queue.enqueue(req1)
        queue.enqueue(req2)

        val first = queue.nextForRun("job/demo#1")
        assertNotNull(first)
        assertEquals("req-1", first.requestId)

        queue.complete("rpc-1")

        val second = queue.nextForRun("job/demo#1")
        assertNotNull(second)
        assertEquals("req-2", second.requestId)
    }
}
