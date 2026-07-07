package auditlogs

type noopAuditLogCollector struct{}

func NewNoopCollector() Collector {
	return new(noopAuditLogCollector)
}

func (*noopAuditLogCollector) CollectMCPAuditEntry(log MCPAuditLog) {}

func (*noopAuditLogCollector) Close() {}
