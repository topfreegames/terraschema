package reader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var localsSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "locals",
		},
	},
}

// staticFunctions are the Terraform functions a local can call and still be evaluated from the module
// source alone, without variables, providers or state.
var staticFunctions = map[string]function.Function{
	"concat":   stdlib.ConcatFunc,
	"contains": stdlib.ContainsFunc,
	"distinct": stdlib.DistinctFunc,
	"flatten":  stdlib.FlattenFunc,
	"keys":     stdlib.KeysFunc,
	"length":   stdlib.LengthFunc,
	"lookup":   stdlib.LookupFunc,
	"lower":    stdlib.LowerFunc,
	"merge":    stdlib.MergeFunc,
	"sort":     stdlib.SortFunc,
	"tolist":   stdlib.MakeToFunc(cty.List(cty.DynamicPseudoType)),
	"tomap":    stdlib.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
	"toset":    stdlib.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
	"upper":    stdlib.UpperFunc,
	"values":   stdlib.ValuesFunc,
}

// GetStaticEvalContext reads all .tf files in a directory and returns an evaluation context holding the
// module's static locals: the ones that evaluate without variables, resources, data sources or modules.
// Validation conditions evaluated in it can then refer to those locals, e.g. contains(keys(local.x), var.y).
func GetStaticEvalContext(path string) (*hcl.EvalContext, error) {
	files, err := filepath.Glob(filepath.Join(path, "*.tf"))
	if err != nil {
		return nil, err
	}

	fileContents := make(map[string][]byte, len(files))
	for _, fileName := range files {
		content, err := os.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %w", fileName, err)
		}
		fileContents[fileName] = content
	}

	return GetStaticEvalContextFromBytes(fileContents)
}

// GetStaticEvalContextFromBytes does what GetStaticEvalContext does for HCL content provided as a map of
// file names to their byte contents.
func GetStaticEvalContextFromBytes(files map[string][]byte) (*hcl.EvalContext, error) {
	parser := hclparse.NewParser()

	pending := make(map[string]hcl.Expression)
	for fileName, content := range files {
		file, d := parser.ParseHCL(content, fileName)
		if d.HasErrors() {
			return nil, d
		}

		blocks, _, d := file.Body.PartialContent(localsSchema)
		if d.HasErrors() {
			return nil, d
		}
		for _, block := range blocks.Blocks {
			attributes, d := block.Body.JustAttributes()
			if d.HasErrors() {
				return nil, d
			}
			for name, attribute := range attributes {
				pending[name] = attribute.Expr
			}
		}
	}

	// Locals can refer to each other in any order, so resolve them in rounds until a round settles nothing.
	resolved := make(map[string]cty.Value)
	for progress := true; progress; {
		progress = false
		for name, expression := range pending {
			ready, static := localDependencies(expression, resolved, pending)
			if static && !ready {
				continue
			}
			delete(pending, name)
			progress = true
			if !static {
				continue
			}

			value, d := expression.Value(staticEvalContext(resolved))
			if !d.HasErrors() && value.IsWhollyKnown() {
				resolved[name] = value
			}
		}
	}

	return staticEvalContext(resolved), nil
}

// localDependencies reports whether an expression only refers to other locals (static) and whether all
// of those are already resolved (ready). A reference to a local that is neither resolved nor pending means
// that local was found not to be static.
func localDependencies(expression hcl.Expression, resolved map[string]cty.Value,
	pending map[string]hcl.Expression,
) (bool, bool) {
	ready := true
	for _, traversal := range expression.Variables() {
		if traversal.RootName() != "local" || len(traversal) < 2 {
			return false, false
		}
		attribute, ok := traversal[1].(hcl.TraverseAttr)
		if !ok {
			return false, false
		}
		if _, ok := resolved[attribute.Name]; ok {
			continue
		}
		if _, ok := pending[attribute.Name]; !ok {
			return false, false
		}
		ready = false
	}

	return ready, true
}

func staticEvalContext(locals map[string]cty.Value) *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{"local": cty.ObjectVal(locals)},
		Functions: staticFunctions,
	}
}
