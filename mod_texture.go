package main

// func ConvertToGLB(
// 	blenderPath string,
// 	scriptPath string,
// 	addonPath string,
// 	vbPath string,
// 	ibPath string,
// 	vbFmt string,
// 	ibFmt string,
// 	outputPath string,
// ) error {
// 	cmd := exec.Command(
// 		blenderPath,
// 		"--background",
// 		"--python",
// 		scriptPath,
// 		"--",
// 		addonPath,
// 		vbPath,
// 		ibPath,
// 		vbFmt,
// 		ibFmt,
// 		outputPath,
// 	)

// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf(
// 			"blender conversion failed: %w\n%s",
// 			err,
// 			output,
// 		)
// 	}

// 	return nil
// }
