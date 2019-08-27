package types

import (
	v1dns "google.golang.org/api/dns/v1"
)

// DNSRecords represents an abstraction for dns record sets collection.
type DNSRecords []*v1dns.ResourceRecordSet
