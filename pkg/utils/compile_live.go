package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	pongo2 "github.com/flosch/pongo2/v4"
	"github.com/spf13/cobra"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/internal/box"
	recurseCopy "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

type CompileLive struct {
}

func (c *CompileLive) ConfirmPackages(sameConfigFile loaders.SameConfig) (map[string]string, error) {
	pipCommand := fmt.Sprintf(`
	#!/bin/bash
	pip3 list --format freeze
		`)

	cmdReturn, err := ExecuteInlineBashScript(&cobra.Command{}, pipCommand, "Pip Freeze failed", false)

	if err != nil {
		log.Tracef("Error executing: %v\n CmdReturn: %v", err.Error(), cmdReturn)
		return map[string]string{}, err
	}
	missingPackages := map[string]string{}
	allPackages := strings.Split(cmdReturn, "\n")
	for _, packageString := range allPackages {
		packageAndVersion := strings.Split(packageString, "==")
		packageName := packageAndVersion[0]
		packageVersion := ""
		if len(packageAndVersion) > 1 {
			packageVersion = packageAndVersion[1]
		}
		missingPackages[packageName] = packageVersion
	}
	// returning nothing for now, figure it out later
	return map[string]string{}, nil
}

func (c *CompileLive) FindAllSteps(convertedText string) (foundSteps []FoundStep, err error) {
	// Need to enable multiline for beginning of the line checking - (?m)
	// Looking for something of the format:
	// # - ...
	// or
	// # + tags=[...]
	re_text := `(?m)^\s*# (?:\+|\-) ?(.*?)$`
	re_steps := regexp.MustCompile(re_text)
	stepsFound := re_steps.FindAllStringSubmatch(convertedText, -1)

	tagsFound := make([][]string, len(stepsFound))
	namedStepsFound := false
	for i, thisStep := range stepsFound {
		tagsFound[i] = ParseTagsForStep(thisStep[1])
		for _, tag := range tagsFound[i] {
			if strings.HasPrefix(tag, "same_step_") {
				namedStepsFound = true
			}
		}
	}

	if !namedStepsFound {
		log.Tracef("no steps found in the file - treating the entire file as a single step.")
		foundStep := FoundStep{}
		foundStep.CodeSlice = convertedText
		foundStep.Index = 0
		foundStep.StepName = "same_step_0"
		foundStep.Tags = nil

		return []FoundStep{foundStep}, nil
	}

	log.Trace("Found at least one step with a 'same_step_#' format, breaking up the file")

	code_blocks_slices := re_steps.Split(convertedText, -1)
	foundSteps = make([]FoundStep, 0)
	current_step_name := "same_step_0"
	current_index := 0
	log.Tracef("Raw steps found: %v", len(stepsFound))
	log.Tracef("Code slices found: %v", len(code_blocks_slices))
	log.Tracef("Raw tag blocks found: %v", len(tagsFound))
	for i := range stepsFound {

		if (i == 0) && (code_blocks_slices[0] == "") {
			// When splitting cells, you can often have a zero cell
			// at the start, so skipping it
			code_blocks_slices = code_blocks_slices[1:]
		}

		cacheValue := ""
		environmentName := ""
		genericTags := make([]string, 0)

		// Drop tags into one  of three categories (should be more extensible in the future)
		for _, tag := range tagsFound[i] {
			if strings.HasPrefix(tag, "same_step_") {
				current_step_name = tag
				current_index, _ = strconv.Atoi(strings.Split(tag, "_")[2])
			} else if strings.HasPrefix(tag, "cache=") {
				cacheValue = strings.Split(tag, "=")[1]
			} else if strings.HasPrefix(tag, "environment=") {
				environmentName = strings.Split(tag, "=")[1]
			} else {
				genericTags = append(genericTags, tag)
			}
		}
		thisFoundStep := FoundStep{}
		thisFoundStep.StepName = current_step_name
		thisFoundStep.CacheValue = cacheValue
		thisFoundStep.EnvironmentName = environmentName
		thisFoundStep.Tags = genericTags
		thisFoundStep.Index = current_index
		thisFoundStep.CodeSlice = code_blocks_slices[i]
		foundSteps = append(foundSteps, thisFoundStep)

	}

	return foundSteps, nil
}

func ParseTagsForStep(s string) []string {
	re_tags_text := `tags=\[([^\]]*)\]`
	re_tags := regexp.MustCompile(re_tags_text)
	tags_found := re_tags.FindAllStringSubmatch(s, -1)
	log.Tracef(" - Tags found: %v\n", len(tags_found))
	if len(tags_found) > 0 {
		all_tags := strings.Split(tags_found[0][1], ",")
		returned_tags := make([]string, len(all_tags))
		for _, this_tag := range all_tags {
			this_tag = strings.TrimSpace(this_tag)
			if this_tag[0] == '"' {
				this_tag = this_tag[1:]
			}
			if end := len(this_tag) - 1; this_tag[end] == '"' {
				this_tag = this_tag[:end]
			}
			returned_tags = append(returned_tags, this_tag)
		}
		return returned_tags
	}

	return nil

}

func (c *CompileLive) CombineCodeSlicesToSteps(foundSteps []FoundStep) (map[string]CodeBlock, error) {
	aggregatedSteps := make(map[string]CodeBlock)
	for _, foundStep := range foundSteps {

		log.Tracef("Current step: %v\n", foundStep.StepName)
		log.Tracef("Current slice: %v\n", foundStep.CodeSlice)

		thisCodeBlock := CodeBlock{}
		if _, exists := aggregatedSteps[foundStep.StepName]; exists {
			thisCodeBlock = aggregatedSteps[foundStep.StepName]
		}

		thisCodeBlock.Code += foundStep.CodeSlice
		thisCodeBlock.StepIdentifier = foundStep.StepName

		if foundStep.CacheValue != "" {
			thisCodeBlock.CacheValue = foundStep.CacheValue
		} else if thisCodeBlock.CacheValue == "" {
			thisCodeBlock.CacheValue = "P0D"
		}

		if foundStep.EnvironmentName != "" {
			thisCodeBlock.EnvironmentName = foundStep.EnvironmentName
		} else if thisCodeBlock.EnvironmentName == "" {
			thisCodeBlock.EnvironmentName = "default"
		}

		aggregatedSteps[foundStep.StepName] = thisCodeBlock
	}

	return aggregatedSteps, nil
}

func (c *CompileLive) CreateRootFile(target string, aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {

	if !ContainsString([]string{"kubeflow", "aml"}, target) {
		return "", fmt.Errorf("unknown compilation target: %v", target)
	}

	rootParameterString := ""

	if len(sameConfigFile.Spec.Run.Parameters) > 0 {
		rootParameters := make(map[string]string, len(sameConfigFile.Spec.Run.Parameters))
		for k, untyped_v := range sameConfigFile.Spec.Run.Parameters {
			switch untyped_v.(type) {
			case int, int8, uint8, int16, uint16, int32, uint32, int64, uint64, uint, uintptr, float32, float64, bool, string:
				rootParameters[k] = fmt.Sprintf("%v", untyped_v)
			default:
				log.Warnf("We only support numeric, bool and strings as default parameters (no dicts or lists). We're setting the default value for '%v' to ''.", k)
				rootParameters[k] = ""
			}

		}
		rootParameterString, _ = JoinMapKeysValues(rootParameters)
	}

	environments := make(map[string]loaders.Environment)
	imagePullSecretsToCreate := make([]loaders.RepositoryCredentials, 0)

	defaultEnvironment := &loaders.Environment{}

	// Pulling from Docker Hub through AML requires the below tag structure of library/name:tag
	defaultEnvironment.ImageTag = "library/python:3.9-slim-buster"
	defaultEnvironment.Packages = make([]string, 0)
	defaultEnvironment.PrivateRegistry = false
	environments["default"] = *defaultEnvironment

	if len(sameConfigFile.Spec.Environments) > 0 {
		for env_name, env := range sameConfigFile.Spec.Environments {
			thisEnvironment := &loaders.Environment{}
			thisEnvironment.ImageTag = ValueOrDefault(env.ImageTag, environments[env_name].ImageTag)
			thisEnvironment.Packages = env.Packages
			thisEnvironment.PrivateRegistry = env.PrivateRegistry
			if thisEnvironment.PrivateRegistry {

				// Two options - either someone has set the secret name (implying it's already mounted, so we'll just move on), or no secret name and so we have to creaate the secret inline.
				// Regardless, we'll just populate this struct, and let the template sort it out
				if thisEnvironment.Credentials.SecretName == "" {
					imagePullSecretsToCreate = append(imagePullSecretsToCreate, env.Credentials)
				}
				thisEnvironment.Credentials = env.Credentials
			}
			environments[env_name] = *thisEnvironment
		}
	}

	previousStep := ""
	stepsLeftToParse := make(map[string]string)
	allSteps := []map[string]string{}
	// Copying this to a new variable so that we can delete them
	for _, thisCodeBlock := range aggregatedSteps {
		stepsLeftToParse[thisCodeBlock.StepIdentifier] = thisCodeBlock.StepIdentifier
	}

	stepsToParse := make([]string, 0, len(stepsLeftToParse))
	for key := range stepsLeftToParse {
		stepsToParse = append(stepsToParse, key)
	}
	sort.Strings(stepsToParse)

	// Unfortunately, every early step's package includes also need to be included in later
	// steps. This is become some objects (like IPython.image) require module imports.
	// There's probably a more elegant way to handle this.
	globalPackagesSlice := make(map[string]string)
	globalPackagesString := ""
	for i := 0; i < len(stepsToParse); i++ {
		thisCodeBlock := aggregatedSteps[stepsToParse[i]]

		// Another hacky work-around - we're just building every package into every container
		// ...definitely should be more efficient (only build in what we need per container)
		packageString := ""
		for k := range thisCodeBlock.PackagesToInstall {
			// Using key mapping on a hash table to eliminate dupes
			globalPackagesSlice[k] = ""
		}

		for k := range globalPackagesSlice {
			packageString += fmt.Sprintf("\"%v\",", k)
			globalPackagesString += fmt.Sprintf("\"%v\",", k)
		}

		imagePullSecretName := ""
		if environments[thisCodeBlock.EnvironmentName].Credentials.SecretName != "" {
			imagePullSecretName = environments[thisCodeBlock.EnvironmentName].Credentials.SecretName
		}

		allSteps = append(allSteps, map[string]string{
			"Name":                thisCodeBlock.StepIdentifier,
			"PackageString":       packageString,
			"CacheValue":          thisCodeBlock.CacheValue,
			"PreviousStep":        previousStep,
			"Environment":         thisCodeBlock.EnvironmentName,
			"ImageName":           environments[thisCodeBlock.EnvironmentName].ImageTag,
			"PrivateRepository":   strconv.FormatBool(environments[thisCodeBlock.EnvironmentName].PrivateRegistry),
			"ImagePullSecretName": imagePullSecretName,
		})

		previousStep = thisCodeBlock.StepIdentifier

	}

	experimentName := removeIllegalExperimentNameCharacters(sameConfigFile.Spec.Metadata.Name)
	stepString := ""
	for _, step := range stepsToParse {
		if stepString != "" {
			stepString += ", "
		}
		stepString += fmt.Sprintf("%v_step", step)
	}

	safeExperimentName := alphaNumericOnly(experimentName)

	rootFileContext := pongo2.Context{
		"RootParameterString":  rootParameterString,
		"GlobalPackagesString": globalPackagesString,
		"Steps":                allSteps,
		"StepString":           stepString,
		"ExperimentName":       experimentName,
		"SafeExperimentName":   safeExperimentName,
		"Kubeconfig":           sameConfigFile.Spec.KubeConfig,
		"Environments":         environments,
		"SecretsToCreate":      imagePullSecretsToCreate,
	}

	var root_file_bytes []byte
	switch target {
	case "kubeflow":
		root_file_bytes = box.Get("/kfp/root.tmpl")
	case "aml":
		root_file_bytes = box.Get("/aml/root.tmpl")
	default:
		return "", fmt.Errorf("unknown compilation target: %v", target)
	}

	tmpl := pongo2.Must(pongo2.FromBytes(root_file_bytes))

	rootFileString, err := tmpl.Execute(rootFileContext)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %v", err)
	}

	return rootFileString, nil

}

func (c *CompileLive) WriteStepFiles(target string, compiledDir string, aggregatedSteps map[string]CodeBlock) (map[string]map[string]string, error) {

	tempStepHolderDir, err := ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	defer os.Remove(tempStepHolderDir)

	returnedPackages := make(map[string]map[string]string)

	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory to write steps to: %v", err)
	}

	for i := range aggregatedSteps {
		returnedPackages[aggregatedSteps[i].StepIdentifier] = make(map[string]string)
		parameterString, _ := JoinMapKeysValues(aggregatedSteps[i].Parameters)
		if parameterString != "" {
			parameterString = "," + parameterString
		}

		// Prepend an empty locals as the default
		parameterString = `__context="gAR9lC4=", __run_info="gAR9lC4=", __metadata_url=""` + parameterString

		stepToWrite := ""
		var step_file_bytes []byte
		switch target {
		case "kubeflow":
			stepToWrite = filepath.Join(compiledDir, fmt.Sprintf("%v.py", aggregatedSteps[i].StepIdentifier))
			step_file_bytes = box.Get("/kfp/step.tmpl")
		case "aml":
			// AML requires each step to be in its own directory, with the same name as the python file
			stepDirectoryName := filepath.Join(compiledDir, aggregatedSteps[i].StepIdentifier)
			_, err := os.Stat(stepDirectoryName)
			if os.IsNotExist(err) {
				errDir := os.MkdirAll(stepDirectoryName, 0700)
				if errDir != nil {
					return nil, fmt.Errorf("error creating step directory for %v: %v", stepDirectoryName, err)
				}

			}

			stepToWrite = filepath.Join(stepDirectoryName, fmt.Sprintf("%v.py", aggregatedSteps[i].StepIdentifier))
			step_file_bytes = box.Get("/aml/step.tmpl")
		default:
			return nil, fmt.Errorf("unknown target: %v", target)
		}

		innerCodeToExecute := ""
		scanner := bufio.NewScanner(strings.NewReader(aggregatedSteps[i].Code))
		for scanner.Scan() {
			innerCodeToExecute += fmt.Sprintln(scanner.Text())
		}

		stepFileContext := pongo2.Context{
			"Name":             aggregatedSteps[i].StepIdentifier,
			"Parameter_String": parameterString,
			"Inner_Code":       innerCodeToExecute,
		}

		tmpl := pongo2.Must(pongo2.FromBytes(step_file_bytes))
		stepFileString, err := tmpl.Execute(stepFileContext)
		if err != nil {
			return nil, fmt.Errorf("error writing step %v: %v", aggregatedSteps[i].StepIdentifier, err.Error())
		}

		err = os.WriteFile(stepToWrite, []byte(stepFileString), 0400)
		if err != nil {
			return nil, fmt.Errorf("Error writing step %v: %v", stepToWrite, err.Error())
		}

		tempStepFile, err := ioutil.TempFile(tempStepHolderDir, fmt.Sprintf("SAME-inner-code-file-*-%v", fmt.Sprintf("%v.py", aggregatedSteps[i].StepIdentifier)))
		if err != nil {
			return nil, fmt.Errorf("error creating tempfile for step %v: %v", aggregatedSteps[i].StepIdentifier, err.Error())
		}

		err = ioutil.WriteFile(tempStepFile.Name(), []byte(innerCodeToExecute), 0400)
		if err != nil {
			return nil, fmt.Errorf("Error writing temporary step file %v: %v", tempStepFile, err.Error())
		}

		log.Tracef("Freezing python packages")
		pipCommand := fmt.Sprintf(`
#!/bin/bash
set -e
pipreqs %v --print 
	`, tempStepHolderDir)

		cmdReturn, err := ExecuteInlineBashScript(&cobra.Command{}, pipCommand, "Pipreqs output failed", false)

		if err != nil {
			log.Tracef("Error executing: %v\n CmdReturn: %v", err.Error(), cmdReturn)
			return nil, err
		}
		allPackages := strings.Split(cmdReturn, "\n")

		for _, packageString := range allPackages {
			if packageString != "" && !strings.HasPrefix(packageString, "INFO: ") {
				returnedPackages[aggregatedSteps[i].StepIdentifier][packageString] = ""
			}
		}

	}

	return returnedPackages, nil
}

func (c *CompileLive) ConvertNotebook(jupytextExecutablePath string, notebookFilePath string) (string, error) {
	log.Infof("Using notebook from here: %v\n", notebookFilePath)
	notebookFile, err := os.Open(notebookFilePath)
	if err != nil {
		return "", fmt.Errorf("program_compile.go: error reading from notebook file: %v", notebookFilePath)
	}

	scriptCmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%v --to py", jupytextExecutablePath))
	scriptStdin, err := scriptCmd.StdinPipe()

	if err != nil {
		return "", fmt.Errorf("Error building Stdin pipe for notebook file: %v", err.Error())
	}

	b, _ := ioutil.ReadAll(notebookFile)

	go func() {
		defer scriptStdin.Close()
		_, _ = io.WriteString(scriptStdin, string(b))
	}()

	out, err := scriptCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error executing notebook conversion: %v", err.Error())
	}

	if err != nil {
		return "", fmt.Errorf(`
could not convert the file: %v
full error message: %v`, notebookFilePath, string(out))
	}

	return string(out), nil
}

func (c CompileLive) WriteSupportFiles(workingDirectory string, directoriesToWriteTo []string) error {
	// Inspired by Kubeflwo - TODO: Make sure to have copyright and license here
	// https://github.com/kubeflow/pipelines/blob/cc83e1089b573256e781ed2e4ac90f604129e769/sdk/python/kfp/containers/_build_image_api.py#L68

	// This function recursively scans the working directory and captures the following files in the container image context:
	// * :code:`requirements.txt` files
	// * All python files

	// Copying all *.py and requirements.txt files

	for _, destDir := range directoriesToWriteTo {
		opt := recurseCopy.Options{
			Skip: func(src string) (bool, error) {
				fi, err := os.Stat(src)
				if err != nil {
					return true, err
				}
				if fi.IsDir() {
					return false, nil
				}
				return !strings.HasSuffix(src, ".py"), nil
			},
			OnDirExists: func(src string, dst string) recurseCopy.DirExistsAction {
				return recurseCopy.Merge
			},
			Sync: true,
		}
		err := recurseCopy.Copy(workingDirectory, destDir, opt)
		if err != nil {
			return fmt.Errorf("Error copying support python files: %v", err)
		}

	}
	return nil
}

func removeIllegalExperimentNameCharacters(s string) string {
	return strings.Map(
		func(r rune) rune {
			if r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		},
		s,
	)
}

func alphaNumericOnly(s string) string {
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(s, ""))
}
