// Copyright 2017 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"fmt"

	"github.com/attic-labs/noms/go/config"
	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/diff"
	"github.com/attic-labs/noms/go/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

func nomsStruct(noms *kingpin.Application) (*kingpin.CmdClause, commandHandler) {
	structCmd := noms.Command("struct", "interface with Struct type values")

	structSetCmd, structSetHandler := structSet(structCmd)

	return structCmd, func(input string) int {
		switch input {
		case structSetCmd.FullCommand():
			return structSetHandler(input)
		}
		d.Panic("notreached")
		return 1
	}
}

func structSet(cmd *kingpin.CmdClause) (*kingpin.CmdClause, commandHandler) {
	structSet := cmd.Command("set", "sets a value in a struct")
	path := structSet.Arg("path", "the path to a struct to set a value in").Required().String()
	key := structSet.Arg("key", "the key to set on the struct").Required().String()
	valInput := structSet.Arg("value", "the value to set the key to").Required().String()

	return structSet, func(input string) int {
		cfg := config.NewResolver()
		// get a spec for this path
		spec, err := cfg.GetSpec(*path)
		if err != nil {
			kingpin.Fatalf("Invalid path %s: %s", *path, err)
			return 1
		}
		// make sure the path describes a dataset so we can commit
		absPath := spec.Path
		if absPath.Dataset == "" {
			// TODO: when db.Flush() exists, remove this and use WriteValue below instead (and flush if dataset exists)
			kingpin.Fatalf("No dataset exists for %s - this command only works with datasets for now", *path)
			return 1
		}
		// get our database and dataset
		db := spec.GetDatabase()
		defer db.Close()
		ds := db.GetDataset(spec.Path.Dataset)
		// make sure there's something to set in
		root, ok := ds.MaybeHead()
		if !ok {
			kingpin.Fatalf("Empty head at %s", *path)
			return 1
		}
		// make sure the specified path exists
		value := spec.GetValue()
		if value == nil {
			kingpin.Fatalf("Invalid path %s", *path)
			return 1
		}
		// make sure the specified key exists
		s := value.(types.Struct)
		prev, ok := s.MaybeGet(*key)
		if !ok {
			kingpin.Fatalf("Key %s does not exist at that path", *key)
		}
		// make sure our specified value is valid
		val, hash, rem, err := types.ParsePathIndex(*valInput)
		if err != nil || rem != "" {
			kingpin.Fatalf("Invalid new value: '%s': %s\n", *valInput, err)
			return 1
		} else if !hash.IsEmpty() {
			// TODO: hash support
			kingpin.Fatalf("Hashes are not supported yet")
			return 1
		}

		// TODO: this is wrong - the value becomes
		// Commit { Commit { value } }
		fmt.Println(prev, val, absPath.Path)

		// assemble a patch with the changes
		difference := diff.Difference{
			Path:       absPath.Path,
			ChangeType: types.DiffChangeModified,
			OldValue:   prev,
			NewValue:   s.Set(*key, val),
		}

		// commit the patch
		newDs, err := db.Commit(ds, diff.Apply(root, diff.Patch{difference}), datas.CommitOptions{})
		if err != nil {
			kingpin.Fatalf("error creating commit: %s", err)
			return 1
		}

		// print the commit hash
		fmt.Print(newDs.HeadRef().Hash())
		return 0
	}
}
