package cloudstack

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

// metadataSchema returns the schema to use for metadata
func metadataSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
	}
}

// setMetadata is a helper to set the metadata for a resource. It expects the
// metadata field to be named "metadata"
func setMetadata(cs *cloudstack.CloudStackClient, d *schema.ResourceData, resourceType string) error {
	if metadata, ok := d.GetOk("metadata"); ok {
		p := cs.Resourcemetadata.NewAddResourceDetailParams(
			tagsFromSchema(metadata.(map[string]interface{})),
			d.Id(), resourceType,
		)
		_, err := cs.Resourcemetadata.AddResourceDetail(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func getMetadata(cs *cloudstack.CloudStackClient, d *schema.ResourceData, resourceType string) (map[string]interface{}, error) {
	p := cs.Resourcemetadata.NewListResourceDetailsParams(resourceType)
	p.SetResourceid(d.Id())
	response, err := cs.Resourcemetadata.ListResourceDetails(p)
	if err != nil {
		return nil, err
	}
	// Only return metadata values that were explicitely set
	var existingFilter map[string]interface{}
	if metadata, ok := d.GetOk("metadata"); ok {
		existingFilter = metadata.(map[string]interface{})
	}
	metadata := make(map[string]interface{}, response.Count)
	for _, detail := range response.ResourceDetails {
		if _, ok := existingFilter[detail.Key]; ok {
			metadata[detail.Key] = detail.Value
		}
	}
	return metadata, nil
}

// updateMetadata is a helper to update only when metadata field change metadata
// field to be named "metadata"
func updateMetadata(cs *cloudstack.CloudStackClient, d *schema.ResourceData, resourceType string) error {
	oraw, nraw := d.GetChange("metadata")
	o := oraw.(map[string]interface{})
	n := nraw.(map[string]interface{})

	remove, create := diffTags(tagsFromSchema(o), tagsFromSchema(n))
	log.Printf("[DEBUG] metadata to remove: %v", remove)
	log.Printf("[DEBUG] metadata to create: %v", create)

	// First remove any obsolete metadata
	if len(remove) > 0 {
		log.Printf("[DEBUG] Removing metadata: %v from %s", remove, d.Id())
		p := cs.Resourcemetadata.NewRemoveResourceDetailParams(d.Id(), resourceType)
		for key := range remove {
			p.SetKey(key)
			_, err := cs.Resourcemetadata.RemoveResourceDetail(p)
			if err != nil {
				return err
			}
		}
	}

	// Then add any new metadata
	if len(create) > 0 {
		log.Printf("[DEBUG] Creating metadata: %v for %s", create, d.Id())
		p := cs.Resourcemetadata.NewAddResourceDetailParams(create, d.Id(), resourceType)
		_, err := cs.Resourcemetadata.AddResourceDetail(p)
		if err != nil {
			return err
		}
	}

	return nil
}
