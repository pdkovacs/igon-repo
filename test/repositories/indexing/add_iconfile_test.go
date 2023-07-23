package indexing

import (
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type addIconfileToIndexTestSuite struct {
	IndexingTestSuite
}

func TestAddIconfileToIndexTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &addIconfileToIndexTestSuite{testSuite})
	}
}

func (s *addIconfileToIndexTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile = icon.Iconfiles[0]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.Error(err)
	s.ErrorIs(err, domain.ErrIconfileAlreadyExists)
}

func (s *addIconfileToIndexTestSuite) TestSecondIconfile() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconfileToIndexTestSuite) TestAddSecondIconfileBySecondUser() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	var secondUser = "sedat"

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile2.IconfileDescriptor, secondUser, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	clone := test_commons.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	s.equalIconAttributes(clone, iconDesc, nil)
}

func (s *addIconfileToIndexTestSuite) TestRollbackOnFailedSideEffect() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	var secondUser = "sedat"

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	cloneOfFirst := test_commons.CloneIcon(icon)

	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile2.IconfileDescriptor, secondUser, func() error {
		return errSideEffectTest
	})
	s.Error(err)
	s.ErrorIs(err, errSideEffectTest)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.equalIconAttributes(cloneOfFirst, iconDesc, nil)
}
