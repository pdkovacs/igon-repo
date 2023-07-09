package api_tests

import (
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authr"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type iconfileDeleteTestSuite struct {
	IconTestSuite
}

func TestIconfileDeleteTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_iconfiledelete") {
		suite.Run(t, &iconfileDeleteTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *iconfileDeleteTestSuite) TestDeletingIconfileFailsWith403WithoutPermission() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, dataIn[0].Iconfiles[1].IconfileDescriptor)

	s.NoError(errDelete)
	s.Equal(403, statusCode)

	resp, descError := session.DescribeAllIcons()
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(dataOut, resp)

	s.AssertEndState()
}

func (s *iconfileDeleteTestSuite) TestDeletingIconfileSucceedsWithRequiredPermission() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, dataIn[0].Iconfiles[1].IconfileDescriptor)
	dataOut[0].Paths = append(dataOut[0].Paths[:1], dataOut[0].Paths[2:]...)

	s.NoError(errDelete)
	s.Equal(204, statusCode)

	resp, descError := session.DescribeAllIcons()
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(dataOut, resp)

	s.AssertEndState()
}

func (s *iconfileDeleteTestSuite) TestDeletingIconfileFailsWith404ForNonexistentIcon() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile("nonexistent-icon", dataIn[0].Iconfiles[1].IconfileDescriptor)

	s.NoError(errDelete)
	s.Equal(404, statusCode)

	resp, descError := session.DescribeAllIcons()
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(dataOut, resp)

	s.AssertEndState()
}

func (s *iconfileDeleteTestSuite) TestDeletingIconfileFailsWith404ForNonexistentIconfile() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	statusCode, errDelete := session.deleteIconfile(dataIn[0].Name, domain.IconfileDescriptor{Format: "nonexistentformat", Size: "18px"})

	s.NoError(errDelete)
	s.Equal(404, statusCode)

	resp, descError := session.DescribeAllIcons()
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(dataOut, resp)

	s.AssertEndState()
}

func (s *iconfileDeleteTestSuite) TestDeleteIconIfLastIconfileDeleted() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_ICONFILE})

	for index := range dataOut[0].Paths {
		statusCode, errDelete := session.deleteIconfile(dataOut[0].Name, dataOut[0].Paths[index].IconfileDescriptor)
		s.NoError(errDelete)
		s.Equal(204, statusCode)
	}

	newDataOut := append(dataOut[:0], dataOut[1:]...)

	resp, descError := session.DescribeAllIcons()
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(newDataOut, resp)

	s.AssertEndState()
}
