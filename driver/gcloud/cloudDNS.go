package gcloud

import (
	"context"
	"io"
	"net/http"

	v1dns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"

	"github.com/trivigy/migrate/v2/types"
)

// CloudDNS represents a driver for gcloud CloudDNS service.
type CloudDNS struct {
	Config
	DNSName     string           `json:"dnsName" yaml:"dnsName"`
	Visibility  string           `json:"visibility" yaml:"visibility"`
	Description string           `json:"description" yaml:"description"`
	Records     types.DNSRecords `json:"records" yaml:"records"`
}

// Create executes the resource creation process.
func (r CloudDNS) Create(out io.Writer) error {
	if err := r.EnsureManagedZone(out); err != nil {
		return err
	}
	if err := r.EnsureRecordSets(out); err != nil {
		return err
	}
	return nil
}

// EnsureManagedZone makes sure that the managed zone is created.
func (r CloudDNS) EnsureManagedZone(out io.Writer) error {
	ctx := context.Background()
	service, err := v1dns.NewService(ctx)
	if err != nil {
		return err
	}

	_, err = service.ManagedZones.Get(r.ProjectID, r.Name).Do()
	if err == nil {
		return nil
	}

	switch err.(*googleapi.Error).Code {
	case http.StatusNotFound:
		_, err := service.ManagedZones.Create(r.ProjectID, &v1dns.ManagedZone{
			Name:        r.Name,
			DnsName:     r.DNSName,
			Visibility:  r.Visibility,
			Description: r.Description,
		}).Do()
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

// EnsureRecordSets makes sure that the needed record sets are created.
func (r CloudDNS) EnsureRecordSets(out io.Writer) error {
	ctx := context.Background()
	service, err := v1dns.NewService(ctx)
	if err != nil {
		return err
	}

	resp, err := service.ResourceRecordSets.List(r.ProjectID, r.Name).Do()
	if err != nil {
		return err
	}

	for _, mustRec := range r.Records {
		additions := []*v1dns.ResourceRecordSet{mustRec}
		var deletions []*v1dns.ResourceRecordSet
		for _, currRec := range resp.Rrsets {
			if currRec.Name == mustRec.Name && currRec.Type == mustRec.Type {
				deletions = append(deletions, currRec)
				break
			}
		}

		_, err = service.Changes.Create(r.ProjectID, r.Name, &v1dns.Change{
			Additions: additions,
			Deletions: deletions,
		}).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

// Destroy executes the resource destruction process.
func (r CloudDNS) Destroy(out io.Writer) error {
	if err := r.DestroyRecordSets(out); err != nil {
		return err
	}
	return nil
}

// DestroyRecordSets deletes the provided records from the cloudDNS service.
func (r CloudDNS) DestroyRecordSets(out io.Writer) error {
	ctx := context.Background()
	service, err := v1dns.NewService(ctx)
	if err != nil {
		return err
	}

	resp, err := service.ResourceRecordSets.List(r.ProjectID, r.Name).Do()
	if err != nil {
		return err
	}

	for _, mustRec := range r.Records {
		var deletions []*v1dns.ResourceRecordSet
		for _, currRec := range resp.Rrsets {
			if currRec.Name == mustRec.Name && currRec.Type == mustRec.Type {
				deletions = append(deletions, currRec)
				break
			}
		}

		_, err = service.Changes.Create(r.ProjectID, r.Name, &v1dns.Change{
			Deletions: deletions,
		}).Do()
		if err != nil {
			return err
		}
	}
	return nil
}
