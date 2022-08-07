package api

import (
	"testing"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authr"
	"igo-repo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type iconfileDeleteTestSuite struct {
	iconTestSuite
}

func TestIconfileDeleteTestSuite(t *testing.T) {
	suite.Run(t, &iconfileDeleteTestSuite{})
}

func (s *iconTestSuite) TestDeletingIconfileFailsWith403WithoutPermission() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, dataIn[0].Iconfiles[1].IconfileDescriptor)

	s.NoError(errDelete)
	s.Equal(403, statusCode)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(dataOut, resp)

	s.assertEndState()
}

func (s *iconTestSuite) TestDeletingIconfileSucceedsWithRequiredPermission() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, dataIn[0].Iconfiles[1].IconfileDescriptor)
	dataOut[0].Paths = append(dataOut[0].Paths[:1], dataOut[0].Paths[2:]...)

	s.NoError(errDelete)
	s.Equal(204, statusCode)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(dataOut, resp)

	s.assertEndState()
}

func (s *iconTestSuite) TestDeletingIconfileFailsWith404ForNonexistentIcon() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile("nonexistent-icon", dataIn[0].Iconfiles[1].IconfileDescriptor)

	s.NoError(errDelete)
	s.Equal(404, statusCode)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(dataOut, resp)

	s.assertEndState()
}

func (s *iconTestSuite) TestDeletingIconfileFailsWith404ForNonexistentIconfile() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, domain.IconfileDescriptor{Format: "nonexistentformat", Size: "18px"})

	s.NoError(errDelete)
	s.Equal(404, statusCode)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(dataOut, resp)

	s.assertEndState()
}

func (s *iconTestSuite) TestDeleteIconIfLastIconfileDeleted() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	for index := range dataOut[0].Paths {
		statusCode, errDelete := session.deleteIconfile(dataOut[0].Name, dataOut[0].Paths[index].IconfileDescriptor)
		s.NoError(errDelete)
		s.Equal(204, statusCode)
	}

	newDataOut := append(dataOut[:0], dataOut[1:]...)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(newDataOut, resp)

	s.assertEndState()
}
