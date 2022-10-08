package repositories_tests

import (
	"testing"

	"igo-repo/internal/app/domain"
	itests_common "igo-repo/test/common"

	"github.com/stretchr/testify/suite"
)

type addTagTestSuite struct {
	DBTestSuite
}

func TestAddTagTestSuite(t *testing.T) {
	suite.Run(t, &addTagTestSuite{})
}

func (s *addTagTestSuite) TestCreateAssociateNonExistingTag() {
	var tags []string
	var err error

	var icon = itests_common.TestData[0]
	const tag = "used-in-marvinjs"

	err = s.dbRepo.CreateIcon(icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	tags, err = s.dbRepo.GetExistingTags()
	s.NoError(err)
	s.Empty(tags)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Empty(iconDesc.Tags)

	err = s.dbRepo.AddTag(icon.Name, tag, icon.ModifiedBy)
	s.NoError(err)

	tags, err = s.dbRepo.GetExistingTags()
	s.NoError(err)
	s.Equal(1, len(tags))
	s.Contains(tags, tag)

	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Tags))
	s.Contains(iconDesc.Tags, tag)
}

func (s *addTagTestSuite) TestReuseExistingTag() {
	var tags []string
	var err error

	var icon1 = itests_common.TestData[0]
	var icon2 = itests_common.TestData[1]
	const tag = "used-in-marvinjs"

	err = s.dbRepo.CreateIcon(icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.CreateIcon(icon2.Name, icon2.Iconfiles[0], icon2.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.AddTag(icon1.Name, tag, icon1.ModifiedBy)
	s.NoError(err)

	tags, err = s.dbRepo.GetExistingTags()
	s.NoError(err)
	s.Equal([]string{tag}, tags)

	var iconDesc1 domain.IconDescriptor
	iconDesc1, err = s.dbRepo.DescribeIcon(icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	var iconDesc2 domain.IconDescriptor
	iconDesc2, err = s.dbRepo.DescribeIcon(icon2.Name)
	s.NoError(err)
	s.Empty(iconDesc2.Tags)

	err = s.dbRepo.AddTag(icon2.Name, tag, icon2.ModifiedBy)
	s.NoError(err)

	iconDesc1, err = s.dbRepo.DescribeIcon(icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	iconDesc2, err = s.dbRepo.DescribeIcon(icon2.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc2.Tags)

	tags, err = s.dbRepo.GetExistingTags()
	s.NoError(err)
	s.Equal([]string{tag}, tags)
}
