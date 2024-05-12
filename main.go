package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	path := "./models"

	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule | packages.NeedDeps}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return
	}

	models := gatherModels(pkgs)

	fmt.Println(models)

}

type model struct {
	ImportPath string
	PkgName    string
	Name       string
}

func gatherModels(pkgs []*packages.Package) []model {
	var models []model
	for _, pkg := range pkgs {
		for k, v := range pkg.TypesInfo.Defs {

			_, ok := v.(*types.TypeName)
			if !ok || !k.IsExported() {
				continue
			}

			if isGORMModel(k.Obj.Decl) {
				models = append(models, model{
					ImportPath: pkg.PkgPath,
					Name:       k.Name,
					PkgName:    pkg.Name,
				})
			}

			objectType := pkg.Types.Scope().Lookup(k.Name)
			viewDefinerInterface := pkg.Types.Scope().Lookup("ViewDefiner").Type().Underlying().(*types.Interface)

			if ok := types.Implements(objectType.Type(), viewDefinerInterface); ok {
				models = append(models, model{
					ImportPath: pkg.PkgPath,
					Name:       k.Name,
					PkgName:    pkg.Name,
				})
			}
		}
	}
	// Return models in deterministic order.
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	return models
}

func isGORMModel(decl any) bool {
	spec, ok := decl.(*ast.TypeSpec)
	if !ok {
		return false
	}
	st, ok := spec.Type.(*ast.StructType)
	if !ok {
		return false
	}
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 && embedsModel(f.Type) {
			return true
		}
	}
	// Look for gorm: tag.
	for _, f := range st.Fields.List {
		if f.Tag == nil {
			continue
		}
		if t := strings.Trim(f.Tag.Value, "`"); reflect.StructTag(t).Get("gorm") != "" {
			return true
		}
	}
	return false
}

// return gorm.Model from the selector expression
func embedsModel(ex ast.Expr) bool {
	s, ok := ex.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == "gorm" && s.Sel.Name == "Model"
}
