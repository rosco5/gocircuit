package source

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
)

type Source struct {
	layout *Layout
	jail   *Jail
	pkg    map[string]*Pkg	// package path to package structure
}

func New(l *Layout, writeDir string) (*Source, error) {
	jail, err := NewJail(writeDir)
	if err != nil {
		return nil, err
	}
	return &Source{
		layout: l,
		jail:   jail,
		pkg:    make(map[string]*Pkg),
	}, nil
}

func (s *Source) GetAll() map[string]*Pkg {
	return s.pkg
}

func (s *Source) GetPkg(pkgPath string) *Pkg {
	return s.pkg[pkgPath]
}

// parses parses package pkg
func (s *Source) ParsePkg(pkgPath string, includeGoRoot bool, mode parser.Mode) (pkg *Pkg, err error) {
	pkgPath = path.Clean(pkgPath)
	
	// Find source root for pkgPath
	var srcDir string
	if srcDir, err = s.layout.FindPkg(pkgPath, includeGoRoot); err != nil {
		return nil, err
	}

	// Save current working directory
	var saveDir string
	if saveDir, err = os.Getwd(); err != nil {
		return nil, err
	}

	// Change current directory to root of sources
	if err = os.Chdir(srcDir); err != nil {
		return nil, err
	}
	defer func() {
		err = os.Chdir(saveDir)
	}()

	// Make file set just for this package
	fset := token.NewFileSet()

	// Parse
	var pkgs map[string]*ast.Package
	if pkgs, err = parser.ParseDir(fset, pkgPath, filterGoNoTest, mode); err != nil {
		return nil, err
	}

	pkg = &Pkg{
		SrcDir:  srcDir,
		FileSet: fset,
		PkgPath: pkgPath,
		PkgAST:  pkgs,
	}
	pkg.link()

	s.pkg[pkgPath] = pkg
	return pkg, nil

}

// TODO: Package source directories will often contain files with main or xxx_test package clauses.
// We ignore those, by guessing they are not part of the program.
// The correct way to ignore is to recognize the comment directive: // +build ignore
func filterGoNoTest(fi os.FileInfo) bool {
	n := fi.Name()
	return len(n) > 0 && strings.HasSuffix(n, ".go") && n[0] != '_' && strings.Index(n, "_test.go") < 0
}
