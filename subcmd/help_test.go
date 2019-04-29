package subcmd

import (
	"strings"
	"testing"

	"github.com/vim-volt/volt/internal/testutil"
)

// 'volt help {cmd}' and 'volt {cmd} -help' output should be same
func TestVoltHelpDiffOutput(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	cmdlist, err := testutil.GetCmdList()
	if err != nil {
		t.Error("testutil.GetCmdList() returned non-nil error: " + err.Error())
	}
	for _, cmd := range cmdlist {
		out1, err := testutil.RunVolt("help", cmd)
		testutil.SuccessExit(t, out1, err)
		out2, err := testutil.RunVolt(cmd, "-help")
		testutil.SuccessExit(t, out2, err)
		if string(out1) != string(out2) {
			t.Errorf("'volt help %s' and 'volt %s -help' output differ", cmd, cmd)
		}
	}
}

// 'volt help {non-existing cmd}' should result in error
func TestErrVoltHelpNonExistingCmd(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("help", "this_cmd_must_not_be_implemented")
	testutil.FailExit(t, out, err)
	if string(out) != "[ERROR] Unknown command 'this_cmd_must_not_be_implemented'\n" {
		t.Error("'volt help {non-existing cmd}' did not show error: " + string(out))
	}
}

// 'volt help help' should show E478
func TestVoltHelpE478(t *testing.T) {
	// =============== setup =============== //

	testutil.SetUpEnv(t)
	defer testutil.CleanUpEnv(t)

	// =============== run =============== //

	out, err := testutil.RunVolt("help", "help")
	testutil.FailExit(t, out, err)
	if !strings.Contains(string(out), "E478: Don't panic!") {
		t.Error("'volt help help' did not show E478 error: " + string(out))
	}
}
