package dynamodb

import (
	"context"
	"fmt"
	"iconrepo/internal/app/domain"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var IconsTableName = "icons"
var iconNameAttribute = "IconName"
var changeIdAttribute = "ChangeId"
var IconTagsTableName = "icon_tags"
var tagAttribute = "Tag"

type DyndbIconfile struct {
	Format string `dynamodbav:"Format"`
	Size   string `dynamodbav:"Size"`
}

func (dyIconfile *DyndbIconfile) toIconfileDescriptor() domain.IconfileDescriptor {
	return domain.IconfileDescriptor{
		Format: dyIconfile.Format,
		Size:   dyIconfile.Size,
	}
}

func (dyIconfile *DyndbIconfile) fromIconfileDescriptor(descriptor domain.IconfileDescriptor) {
	newIconfile := DyndbIconfile{
		Format: descriptor.Format,
		Size:   descriptor.Size,
	}
	*dyIconfile = newIconfile
}

func toIconfileDescriptorList(iconfiles []DyndbIconfile) []domain.IconfileDescriptor {
	descList := []domain.IconfileDescriptor{}
	for _, iconfile := range iconfiles {
		descList = append(descList, iconfile.toIconfileDescriptor())
	}
	return descList
}

type DyndbIcon struct {
	IconName   string          `dynamodbav:"IconName"`
	ChangeId   string          `dynamodbav:"ChangeId"`
	ModifiedBy string          `dynamodbav:"ModifiedBy"`
	Iconfiles  []DyndbIconfile `dynamodbav:"Iconfiles"`
	Tags       []string        `dynamodbav:"Tags"`
}

func (dyIcon *DyndbIcon) GetKey(ctx context.Context) (map[string]types.AttributeValue, error) {
	iconName, err := attributevalue.Marshal(dyIcon.IconName)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dynamodb attribute `IconName`: %w", Unwrap(ctx, err))
	}
	return map[string]types.AttributeValue{iconNameAttribute: iconName}, nil
}

func (dyIcon *DyndbIcon) unmarshal(attrmap map[string]types.AttributeValue) error {
	unmarshalErr := attributevalue.UnmarshalMap(attrmap, dyIcon)
	if unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal %T: %w", DyndbIcon{}, unmarshalErr)
	}
	if dyIcon.Tags == nil {
		dyIcon.Tags = []string{}
	}

	return nil
}

func (dyIcon *DyndbIcon) toIconDescriptor() domain.IconDescriptor {
	return domain.IconDescriptor{
		IconAttributes: domain.IconAttributes{
			Name:       dyIcon.IconName,
			ModifiedBy: dyIcon.ModifiedBy,
			Tags:       dyIcon.Tags,
		},
		Iconfiles: toIconfileDescriptorList(dyIcon.Iconfiles),
	}
}

type DyndbTag struct {
	Tag            string `dynamodbav:"Tag"`
	ReferenceCount int64  `dynamodbav:"ReferenceCount"`
	ChangeId       string `dynamodbav:"ChangeId"`
}

func (dyTag *DyndbTag) GetKey(ctx context.Context) (map[string]types.AttributeValue, error) {
	tagAttributeValue, err := attributevalue.Marshal(dyTag.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get key for %v: %w", dyTag.Tag, Unwrap(ctx, err))
	}
	return map[string]types.AttributeValue{
		tagAttribute: tagAttributeValue,
	}, nil
}

func (dyTag *DyndbTag) unmarshal(attrmap map[string]types.AttributeValue) error {
	unmarshalErr := attributevalue.UnmarshalMap(attrmap, dyTag)
	if unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal %T: %w", DyndbTag{}, unmarshalErr)
	}

	return nil
}
