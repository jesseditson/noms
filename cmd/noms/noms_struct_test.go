// Copyright 2017 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"fmt"
	"testing"

	"github.com/attic-labs/noms/go/spec"
	"github.com/attic-labs/noms/go/types"
	"github.com/attic-labs/noms/go/util/clienttest"
	"github.com/attic-labs/testify/suite"
)

func TestNomsStruct(t *testing.T) {
	suite.Run(t, &nstrSuite{})
}

type nstrSuite struct {
	clienttest.ClientTestSuite
}

func (s *nstrSuite) TestNomsStructSet() {
	sp, err := spec.ForDatabase(s.TempDir)
	s.NoError(err)
	defer sp.Close()
	db := sp.GetDatabase()
	testValue := types.NewStruct("Test", types.StructData{
		"foo": types.String("bar"),
	})
	_, err = db.CommitValue(db.GetDataset("datasetID"), testValue)
	s.NoError(err)

	// TODO: is it right that this is .value?
	path := fmt.Sprintf("%s::datasetID.value", s.TempDir)

	stdout, _ := s.MustRun(main, []string{"struct", "set", path, "foo", "\"not-bar\""})
	// outputs the hash
	s.Equal(stdout, "motbeva9dndv4k19tdhepeodvfnrahfo")

	// hash is now head
	db.Rebase()
	s.Equal(stdout, db.GetDataset("datasetID").HeadRef().Hash().String())

	sh, e := s.MustRun(main, []string{"show", fmt.Sprintf("%s::datasetID", s.TempDir)})
	fmt.Println(sh, e)
}

// test creating new values

// test set on a hash
