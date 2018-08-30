package gparser

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type (
		in struct {
			query string
		}
		out struct {
			node Node
			err  error
		}
	)
	cases := []struct {
		name string
		in   in
		out  out
	}{
		{
			name: "empty",
			in: in{
				query: ``,
			},
			out: out{
				node: new(EmptyNode),
			},
		},
		{
			name: "only value",
			in: in{
				query: `foo`,
			},
			out: out{
				node: &ValueNode{Value: `foo`},
			},
		},
		{
			name: "complete match value",
			in: in{
				query: `"foo bar"`,
			},
			out: out{
				node: &CompleteMatchNode{Value: `foo bar`},
			},
		},
		{
			name: "and query",
			in: in{
				query: `foo bar`,
			},
			out: out{
				node: &AndNode{
					Left:  &ValueNode{Value: `foo`},
					Right: &ValueNode{Value: `bar`},
				},
			},
		},
		{
			name: "and query with operator",
			in: in{
				query: `foo and bar`,
			},
			out: out{
				node: &AndNode{
					Left:  &ValueNode{Value: `foo`},
					Right: &ValueNode{Value: `bar`},
				},
			},
		},
		{
			name: "multiple and qury",
			in: in{
				query: `foo and bar "foo bar"`,
			},
			out: out{
				node: &AndNode{
					Left: &AndNode{
						Left:  &ValueNode{Value: `foo`},
						Right: &ValueNode{Value: `bar`},
					},
					Right: &CompleteMatchNode{Value: `foo bar`},
				},
			},
		},
		{
			name: "or query",
			in: in{
				query: `foo or bar`,
			},
			out: out{
				node: &OrNode{
					Left:  &ValueNode{Value: `foo`},
					Right: &ValueNode{Value: `bar`},
				},
			},
		},
		{
			name: "multiple or query",
			in: in{
				query: `foo or bar or "foo bar"`,
			},
			out: out{
				node: &OrNode{
					Left: &OrNode{
						Left:  &ValueNode{Value: `foo`},
						Right: &ValueNode{Value: `bar`},
					},
					Right: &CompleteMatchNode{Value: `foo bar`},
				},
			},
		},
		{
			name: "keys query",
			in: in{
				query: `foo:bar`,
			},
			out: out{
				node: &KeyNode{
					Key:   `foo`,
					Value: &ValueNode{Value: `bar`},
				},
			},
		},
		{
			name: "grouped query",
			in: in{
				query: `foo (bar or foobar)`,
			},
			out: out{
				node: &AndNode{
					Left: &ValueNode{Value: `foo`},
					Right: &OrNode{
						Left:  &ValueNode{Value: `bar`},
						Right: &ValueNode{Value: `foobar`},
					},
				},
			},
		},
		{
			name: "comprex query",
			in: in{
				query: `foo and (key:value or (foobar and key2:"value2")) or key3:value3 FOO`,
			},
			out: out{
				node: &AndNode{
					Left: &OrNode{
						Left: &AndNode{
							Left: &ValueNode{Value: `foo`},
							Right: &OrNode{
								Left: &KeyNode{
									Key:   `key`,
									Value: &ValueNode{Value: `value`},
								},
								Right: &AndNode{
									Left: &ValueNode{Value: `foobar`},
									Right: &KeyNode{
										Key:   `key2`,
										Value: &CompleteMatchNode{Value: `value2`},
									},
								},
							},
						},
						Right: &KeyNode{
							Key:   `key3`,
							Value: &ValueNode{Value: `value3`},
						},
					},
					Right: &ValueNode{Value: `FOO`},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in.query)
			if a, e := err, c.out.err; !reflect.DeepEqual(a, e) {
				t.Errorf("return error\nwant: %#v\n got: %#v", e, a)
			}
			if a, e := got, c.out.node; !reflect.DeepEqual(a, e) {
				t.Errorf("Nodes\nwant: %v\n got: %v", e, a)
			}
		})
	}
}
