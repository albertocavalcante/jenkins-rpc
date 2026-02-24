package io.albertocavalcante.jenkins.steprpc

import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ConcurrentLinkedQueue

data class PendingBridgeRequest(
    val requestId: String,
    val runId: String,
    val operation: String,
    val args: Map<String, Any?>,
    val targetRunExternalizableId: String,
)

class CpsBridgeQueue {
    private val byTargetRun = ConcurrentHashMap<String, ConcurrentLinkedQueue<PendingBridgeRequest>>()
    private val targetRunByRunID = ConcurrentHashMap<String, String>()

    fun enqueue(request: PendingBridgeRequest) {
        val queue = byTargetRun.computeIfAbsent(request.targetRunExternalizableId) { ConcurrentLinkedQueue() }
        queue.add(request)
        targetRunByRunID[request.runId] = request.targetRunExternalizableId
    }

    fun nextForRun(targetRunExternalizableId: String): PendingBridgeRequest? {
        return byTargetRun[targetRunExternalizableId]?.peek()
    }

    fun complete(runId: String): PendingBridgeRequest? {
        val targetRun = targetRunByRunID.remove(runId) ?: return null
        val queue = byTargetRun[targetRun] ?: return null
        val match = queue.firstOrNull { it.runId == runId } ?: return null
        queue.remove(match)
        if (queue.isEmpty()) {
            byTargetRun.remove(targetRun, queue)
        }
        return match
    }
}
