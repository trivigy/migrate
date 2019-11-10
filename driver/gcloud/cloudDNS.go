package gcloud

import (
	"context"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	v1dns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"

	"github.com/trivigy/migrate/v2/types"
)

// CloudDNS represents a driver for gcloud CloudDNS service.
type CloudDNS struct {
	Profile
	DNSName     string           `json:"dnsName" yaml:"dnsName" validate:"required"`
	Visibility  string           `json:"visibility" yaml:"visibility" validate:"required"`
	Description string           `json:"description" yaml:"description" validate:"required"`
	Records     types.DNSRecords `json:"records" yaml:"records"`
}

var _ interface {
	types.Creator
	types.Destroyer
} = new(CloudDNS)

// Create executes the resource creation process.
func (r CloudDNS) Create(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.EnsureManagedZone(ctx, out, ts); err != nil {
		return err
	}
	if err := r.EnsureRecordSets(ctx, out, ts); err != nil {
		return err
	}
	return nil
}

// EnsureManagedZone makes sure that the managed zone is created.
func (r CloudDNS) EnsureManagedZone(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := v1dns.NewService(ctx, option.WithTokenSource(ts))
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
func (r CloudDNS) EnsureRecordSets(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := v1dns.NewService(ctx, option.WithTokenSource(ts))
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
func (r CloudDNS) Destroy(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.DestroyRecordSets(ctx, out, ts); err != nil {
		return err
	}
	return nil
}

// DestroyRecordSets deletes the provided records from the cloudDNS service.
func (r CloudDNS) DestroyRecordSets(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := v1dns.NewService(ctx, option.WithTokenSource(ts))
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

		if len(deletions) != 0 {
			_, err = service.Changes.Create(r.ProjectID, r.Name, &v1dns.Change{
				Deletions: deletions,
			}).Do()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
