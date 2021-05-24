package repositories

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/itests"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addTagTestSuite struct {
	dbTestSuite
}

func TestAddTagTestSuite(t *testing.T) {
	suite.Run(t, &addTagTestSuite{})
}

func (s *addTagTestSuite) TestCreateAssociateNonExistingTag() {
	var tags []string
	var err error

	var icon = itests.TestData[0]
	const tag = "used-in-marvinjs"

	err = repositories.CreateIcon(getPool(), icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	tags, err = repositories.GetExistingTags(getPool())
	s.NoError(err)
	s.Empty(tags)

	var iconDesc domain.Icon
	iconDesc, err = repositories.DescribeIcon(db, icon.Name)
	s.NoError(err)
	s.Empty(iconDesc.Tags)

	err = repositories.AddTag(db, icon.Name, tag, icon.ModifiedBy)
	s.NoError(err)

	tags, err = repositories.GetExistingTags(getPool())
	s.NoError(err)
	s.Equal(1, len(tags))
	s.Contains(tags, tag)

	iconDesc, err = repositories.DescribeIcon(db, icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Tags))
	s.Contains(iconDesc.Tags, tag)
}

func (s *addTagTestSuite) TestReuseExistingTag() {
	var tags []string
	var err error

	var icon1 = itests.TestData[0]
	var icon2 = itests.TestData[1]
	const tag = "used-in-marvinjs"

	err = repositories.CreateIcon(getPool(), icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.CreateIcon(getPool(), icon2.Name, icon2.Iconfiles[0], icon2.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.AddTag(db, icon1.Name, tag, icon1.ModifiedBy)
	s.NoError(err)

	tags, err = repositories.GetExistingTags(getPool())
	s.NoError(err)
	s.Equal([]string{tag}, tags)

	var iconDesc1 domain.Icon
	iconDesc1, err = repositories.DescribeIcon(db, icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	var iconDesc2 domain.Icon
	iconDesc2, err = repositories.DescribeIcon(db, icon2.Name)
	s.NoError(err)
	s.Empty(iconDesc2.Tags)

	err = repositories.AddTag(db, icon2.Name, tag, icon2.ModifiedBy)
	s.NoError(err)

	iconDesc1, err = repositories.DescribeIcon(db, icon1.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc1.Tags)

	iconDesc2, err = repositories.DescribeIcon(db, icon2.Name)
	s.NoError(err)
	s.Equal([]string{tag}, iconDesc2.Tags)

	tags, err = repositories.GetExistingTags(getPool())
	s.NoError(err)
	s.Equal([]string{tag}, tags)
}
