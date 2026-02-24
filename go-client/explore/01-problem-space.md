# 01 - Problem Space

## Goal

Define a robust Go client contract for invoking Jenkins step-RPC operations and tracking execution state.

## Problem

The client must interact with a plugin that queues work into Jenkins runtime context. This requires explicit handling for auth, retries, polling, and schema stability.

## Constraints

1. Jenkins endpoints may return asynchronous states.
2. Transport errors and terminal execution errors must be distinct.
3. Client behavior must support cancellation and deadlines.
4. Backward compatibility is required across plugin updates.
