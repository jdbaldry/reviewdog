package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/reviewdog/reviewdog/proto/rdf"
	"google.golang.org/protobuf/encoding/protojson"
)

func ExampleDiffParser() {
	const sample = `diff --git a/gofmt.go b/gofmt.go
--- a/gofmt.go	2020-07-26 08:01:09.260800318 +0000
+++ b/gofmt.go	2020-07-26 08:01:09.260800318 +0000
@@ -1,6 +1,6 @@
 package testdata
 
-func    fmt     () {
+func fmt() {
 	// test
 	// test line
 	// test line
@@ -10,11 +10,11 @@
 	// test line
 	// test line
 
-println(
-		"hello, gofmt test"    )
-//comment
+	println(
+		"hello, gofmt test")
+	//comment
 }
 
+type s struct{ A int }
 
-type s struct { A int }
 func (s s) String() { return "s" }
`
	const strip = 1
	p := NewDiffParser(strip)
	diagnostics, err := p.Parse(strings.NewReader(sample))
	if err != nil {
		panic(err)
	}
	for _, d := range diagnostics {
		rdjson, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(d)
		var out bytes.Buffer
		json.Indent(&out, rdjson, "", "  ")
		fmt.Println(out.String())
	}
	// Output:
	// {
	//   "location": {
	//     "path": "gofmt.go",
	//     "range": {
	//       "start": {
	//         "line": 3
	//       },
	//       "end": {
	//         "line": 3
	//       }
	//     }
	//   },
	//   "suggestions": [
	//     {
	//       "range": {
	//         "start": {
	//           "line": 3
	//         },
	//         "end": {
	//           "line": 3
	//         }
	//       },
	//       "text": "func fmt() {"
	//     }
	//   ],
	//   "originalOutput": "gofmt.go:3:-func    fmt     () {\ngofmt.go:3:+func fmt() {"
	// }
	// {
	//   "location": {
	//     "path": "gofmt.go",
	//     "range": {
	//       "start": {
	//         "line": 13
	//       },
	//       "end": {
	//         "line": 15
	//       }
	//     }
	//   },
	//   "suggestions": [
	//     {
	//       "range": {
	//         "start": {
	//           "line": 13
	//         },
	//         "end": {
	//           "line": 15
	//         }
	//       },
	//       "text": "\tprintln(\n\t\t\"hello, gofmt test\")\n\t//comment"
	//     }
	//   ],
	//   "originalOutput": "gofmt.go:13:-println(\ngofmt.go:14:-\t\t\"hello, gofmt test\"    )\ngofmt.go:15:-//comment\ngofmt.go:13:+\tprintln(\ngofmt.go:14:+\t\t\"hello, gofmt test\")\ngofmt.go:15:+\t//comment"
	// }
	// {
	//   "location": {
	//     "path": "gofmt.go",
	//     "range": {
	//       "start": {
	//         "line": 18,
	//         "column": 1
	//       },
	//       "end": {
	//         "line": 18,
	//         "column": 1
	//       }
	//     }
	//   },
	//   "suggestions": [
	//     {
	//       "range": {
	//         "start": {
	//           "line": 18,
	//           "column": 1
	//         },
	//         "end": {
	//           "line": 18,
	//           "column": 1
	//         }
	//       },
	//       "text": "type s struct{ A int }\n"
	//     }
	//   ],
	//   "originalOutput": "gofmt.go:18:+type s struct{ A int }"
	// }
	// {
	//   "location": {
	//     "path": "gofmt.go",
	//     "range": {
	//       "start": {
	//         "line": 19
	//       },
	//       "end": {
	//         "line": 19
	//       }
	//     }
	//   },
	//   "suggestions": [
	//     {
	//       "range": {
	//         "start": {
	//           "line": 19
	//         },
	//         "end": {
	//           "line": 19
	//         }
	//       }
	//     }
	//   ],
	//   "originalOutput": "gofmt.go:19:-type s struct { A int }"
	// }
}

func TestAddTrailingNewlineSuggestion(t *testing.T) {
	t.Parallel()

	const addNewlineDiff = `diff --git a/wantnewline.txt b/wantnewline.txt
index ea74bcd..fd772f0 100644
--- a/wantnewline.txt
+++ b/wantnewline.txt
@@ -1 +1 @@
-No newline at end of the old file but it is present in the new file
\ No newline at end of file
+No newline at end of the old file but it is present in the new file
`

	const strip = 1

	p := NewDiffParser(strip)
	want := []*rdf.Diagnostic{
		{
			Location: &rdf.Location{
				Path: "wantnewline.txt",
				Range: &rdf.Range{
					Start: &rdf.Position{
						Line:   1,
						Column: 0,
					},
					End: &rdf.Position{
						Line:   1,
						Column: 0,
					},
				},
			},
			Suggestions: []*rdf.Suggestion{
				{
					Text: "No newline at end of the old file but it is present in the new file\n",
					Range: &rdf.Range{
						Start: &rdf.Position{
							Line:   1,
							Column: 0,
						},
						End: &rdf.Position{
							Line:   1,
							Column: 0,
						},
					},
				},
			},
			OriginalOutput: "wantnewline.txt:1:-No newline at end of the old file but it is present in the new file\nwantnewline.txt:1:+No newline at end of the old file but it is present in the new file\n",
		},
	}

	got, err := p.Parse(strings.NewReader(addNewlineDiff))
	if err != nil {
		t.Fatalf("%s unexpected error: %v", t.Name(), err)
	}

	ignoreUnexported := cmpopts.IgnoreUnexported(rdf.Diagnostic{}, rdf.Location{}, rdf.Range{}, rdf.Position{}, rdf.Suggestion{})

	if diff := cmp.Diff(want, got, ignoreUnexported); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", t.Name(), diff)
	}
}
