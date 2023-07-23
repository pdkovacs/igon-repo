package indexing

import (
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type addTagTestSuite struct {
	IndexingTestSuite
}

func TestAddTagTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &addTagTestSuite{testSuite})
	}
}

func (s *addTagTestSuite) TestCreateAssociateNonExistingTag() {
	var tags []string
	var err error

	var icon = test_commons.TestData[0]
	const tag = "used-in-marvinjs"

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	tags, err = s.testRepoController.GetExistingTags(s.ctx)
	s.NoError(err)
	s.Empty(tags)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.Empty(iconDesc.Tags)

	err = s.testRepoController.AddTag(s.ctx, icon.Name, tag, icon.ModifiedBy)
	s.NoError(err)

	tags, err = s.testRepoController.GetExistingTags(s.ctx)
	s.NoError(err)
	s.Equal(1, len(tags))
	s.Contains(tags, tag)

	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Tags))
	s.Contains(iconDesc.Tags, tag)
}

func (s *addTagTestSuite) TestReuseExistingTag() {
	var tags []string
	var err error

	var icon1 = test_commons.TestData[0]
	var icon2 = test_commons.TestData[1]
	const tag = "used-in-marvinjs"

	err = s.testRepoController.CreateIcon(s.ctx, icon1.Name, icon1.Iconfiles[0].IconfileDescriptor, icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.CreateIcon(s.ctx, icon2.Name, icon2.Iconfiles[0].IconfileDescriptor, icon2.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddTag(s.ctx, icon1.Name, tag, icon1.ModifiedBy)
	s.NoError(err)

	tags, err = s.testRepoController.GetExistingTags(s.ctx)
	s.NoError(err)
	s.Equal([]string{tag}, tags)

	var iconDesc1 domain.IconDescriptor
	iconDesc1, err = s.testRepoController.DescribeIcon(s.ctx, icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	var iconDesc2 domain.IconDescriptor
	iconDesc2, err = s.testRepoController.DescribeIcon(s.ctx, icon2.Name)
	s.NoError(err)
	s.Empty(iconDesc2.Tags)

	err = s.testRepoController.AddTag(s.ctx, icon2.Name, tag, icon2.ModifiedBy)
	s.NoError(err)

	iconDesc1, err = s.testRepoController.DescribeIcon(s.ctx, icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	iconDesc2, err = s.testRepoController.DescribeIcon(s.ctx, icon2.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc2.Tags)

	tags, err = s.testRepoController.GetExistingTags(s.ctx)
	s.NoError(err)
	s.Equal([]string{tag}, tags)
}
