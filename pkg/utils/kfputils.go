package utils

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func CompileForKFP(pipelineDSLfilepath string) (compiledPipeline string, err error) {

	if _, err = exec.LookPath("dsl-compile"); err != nil {
		err = fmt.Errorf("could not find 'dsl-compile'. Please run 'python -m pip install kfp'. You may also need to add it to your path by executing: export PATH=$PATH:$HOME/.local/bin")
		return
	}

	log.Tracef("Pipeline DSL filepath: %v", pipelineDSLfilepath)
	pipelineDSLfilepath, err = ResolveLocalFilePath(pipelineDSLfilepath)
	if err != nil {
		err = fmt.Errorf("kfputils.go: could not find pipeline definition specified in SAME program: %v", pipelineDSLfilepath)
		return
	}

	tmpfile, _ := ioutil.TempFile("", "SAME-*.tar.gz")
	compiledPipeline = tmpfile.Name()

	scriptCmd := exec.Command("dsl-compile", "--py", pipelineDSLfilepath, "--output", compiledPipeline)
	log.Tracef("About to execute: %v", scriptCmd)
	out, err := scriptCmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf(`
could not compile pipeline: %v
dsl-compile error message: %v`, pipelineDSLfilepath, string(out))
	}
	return
}
