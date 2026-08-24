# Metering

Consumes `usage.recorded` from JetStream using a durable pull consumer.

Important design property: ACK happens only after the business handler succeeds.

Next milestone: persist per-period usage counters and event IDs in Couchbase so redelivery cannot double count usage.
