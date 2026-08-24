# Contract tests

Regenerate protobuf code from `proto/wallet.proto` and test the gRPC service against the generated client.

Keep protobuf changes backwards compatible: do not reuse field numbers and prefer additive changes.
