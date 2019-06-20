package bootstrap

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

func Interactive() {
	/*
		- enter project dir
		- enter cluster name
		- do you want to import helm repositories installed in the system?
		- do you want to import releases from the helmfile
	*/

	prompt := promptui.Prompt{
		Label:   "Please enter the path for your project ",
		Default: ".",
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", result)

	prompt = promptui.Prompt{
		Label:   "Name of the cluster used by default, used to populate directory tree ",
		Default: "dev",
	}
	result, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", result)

	prompt = promptui.Prompt{
		Label:     "Do you want to import helm repositories installed in the system? ",
		IsConfirm: true,
	}
	result, err = prompt.Run()
	if err != nil && err != promptui.ErrAbort {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", result)

	prompt = promptui.Prompt{
		Label:     "Do you want to import releases from the helmfile? ",
		IsConfirm: true,
	}
	result, err = prompt.Run()
	if err != nil && err != promptui.ErrAbort {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", result)
	if result == "y" {
		prompt = promptui.Prompt{
			Label:   "Please enter the path to your helmfile",
			Default: "./helmfile.yaml",
		}
		result, err = prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		fmt.Printf("You choose %q\n", result)
	}
}
