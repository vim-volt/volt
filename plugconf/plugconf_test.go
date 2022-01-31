package plugconf

import (
	"reflect"
	"testing"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
)

func TestSortByDepends(t *testing.T) {
	type input struct {
		reposList   []lockjson.Repos
		plugconfMap map[pathutil.ReposPath]*ParsedInfo
	}

	cases := []struct {
		input input
		want  []lockjson.Repos
	}{
		{
			input: input{
				reposList: []lockjson.Repos{
					{Path: pathutil.DecodeReposPath("test/test-1")},
					{Path: pathutil.DecodeReposPath("test/test-2")},
					{Path: pathutil.DecodeReposPath("test/test-3")},
				},
				plugconfMap: map[pathutil.ReposPath]*ParsedInfo{
					pathutil.DecodeReposPath("test/test-1"): {
						depends: []pathutil.ReposPath{
							pathutil.DecodeReposPath("test/test-2"),
						},
					},
					pathutil.DecodeReposPath("test/test-2"): {
						depends: []pathutil.ReposPath{
							pathutil.DecodeReposPath("test/test-3"),
						},
					},
					pathutil.DecodeReposPath("test/test-3"): {},
				},
			},
			want: []lockjson.Repos{
				{Path: pathutil.DecodeReposPath("test/test-3")},
				{Path: pathutil.DecodeReposPath("test/test-2")},
				{Path: pathutil.DecodeReposPath("test/test-1")},
			},
		},
		{
			input: input{
				reposList: []lockjson.Repos{
					{Path: pathutil.DecodeReposPath("Shougo/ddc-matcher_head")},
					{Path: pathutil.DecodeReposPath("Shougo/ddc.vim")},
					{Path: pathutil.DecodeReposPath("shun/ddc-vim-lsp")},
					{Path: pathutil.DecodeReposPath("vim-denops/denops.vim")},
				},
				plugconfMap: map[pathutil.ReposPath]*ParsedInfo{
					pathutil.DecodeReposPath("vim-denops/denops.vim"): {},
					pathutil.DecodeReposPath("Shougo/ddc.vim"): {
						depends: []pathutil.ReposPath{
							pathutil.DecodeReposPath("vim-denops/denops.vim"),
						},
					},
					pathutil.DecodeReposPath("Shougo/ddc-matcher_head"): {
						depends: []pathutil.ReposPath{
							pathutil.DecodeReposPath("vim-denops/denops.vim"),
							pathutil.DecodeReposPath("Shougo/ddc.vim"),
						},
					},
					pathutil.DecodeReposPath("shun/ddc-vim-lsp"): {
						depends: []pathutil.ReposPath{
							pathutil.DecodeReposPath("vim-denops/denops.vim"),
							pathutil.DecodeReposPath("Shougo/ddc.vim"),
							pathutil.DecodeReposPath("Shougo/ddc-matcher_head"),
						},
					},
				},
			},
			want: []lockjson.Repos{
				{Path: pathutil.DecodeReposPath("vim-denops/denops.vim")},
				{Path: pathutil.DecodeReposPath("Shougo/ddc.vim")},
				{Path: pathutil.DecodeReposPath("Shougo/ddc-matcher_head")},
				{Path: pathutil.DecodeReposPath("shun/ddc-vim-lsp")},
			},
		},
	}

	for _, tt := range cases {
		sortByDepends(tt.input.reposList, tt.input.plugconfMap)

		if !reflect.DeepEqual(tt.input.reposList, tt.want) {
			t.Fatalf("want: %v, but got: %v", tt.want, tt.input.reposList)
		}
	}
}
