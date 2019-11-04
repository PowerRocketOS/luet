// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

// Recipe is a builder imeplementation.

// It reads a Tree and spit it in human readable form (YAML), called recipe,
// It also loads a tree (recipe) from a YAML (to a db, e.g. BoltDB), allowing to query it
// with the solver, using the package object.
package tree

import (
	"io/ioutil"
	"os"
	"path/filepath"

	pkg "github.com/mudler/luet/pkg/package"
)

const (
	DefinitionFile = "definition.yaml"
)

func NewGeneralRecipe() Builder { return &Recipe{} }

// Recipe is the "general" reciper for Trees
type Recipe struct {
	PackageTree pkg.Tree
}

func (r *Recipe) Save(path string) error {

	for _, pid := range r.PackageTree.GetPackageSet().GetPackages() {

		p, err := r.PackageTree.GetPackageSet().GetPackage(pid)
		if err != nil {
			return err
		}
		dir := filepath.Join(path, p.GetCategory(), p.GetName(), p.GetVersion())
		os.MkdirAll(dir, os.ModePerm)
		data, err := p.Yaml()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dir, DefinitionFile), data, 0644)
		if err != nil {
			return err
		}

	}
	return nil
}

func (r *Recipe) Load(path string) error {

	if r.Tree() == nil {
		r.PackageTree = NewDefaultTree()
	}

	tmpfile, err := ioutil.TempFile("", "luet")
	if err != nil {
		return err
	}

	r.Tree().SetPackageSet(pkg.NewBoltDatabase(tmpfile.Name()))
	// TODO: Handle cleaning after? Cleanup implemented in GetPackageSet().Clean()

	// the function that handles each file or dir
	var ff = func(currentpath string, info os.FileInfo, err error) error {

		if info.Name() != DefinitionFile {
			return nil // Skip with no errors
		}

		dat, err := ioutil.ReadFile(currentpath)
		if err != nil {
			return err
		}
		pack, err := pkg.DefaultPackageFromYaml(dat)
		if err != nil {
			return err
		}

		pack.SetPath(filepath.Dir(path))

		_, err = r.Tree().GetPackageSet().CreatePackage(&pack)
		if err != nil {
			return err
		}
		// first thing to do, check error. and decide what to do about it
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(path, ff)
	if err != nil {
		return err
	}
	return nil
}

func (r *Recipe) Tree() pkg.Tree      { return r.PackageTree }
func (r *Recipe) WithTree(t pkg.Tree) { r.PackageTree = t }