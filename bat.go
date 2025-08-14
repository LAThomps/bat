package main

import (
	"fmt"
	"os"
	"strconv"
	"io"
	"errors"
)

const batFiles string = "/sys/class/power_supply/BAT0/"
const startThresh string = batFiles + "charge_control_start_threshold"
const endThresh string = batFiles + "charge_control_end_threshold"
const capacityLevel string = batFiles + "capacity"

func main() {
	// for i, arg := range os.Args {
	// 	fmt.Printf("Argument %d: %s\n", i, arg)
	// }
	start_old, end_old, capacity := read_current_levels()
	if len(os.Args) == 1 {
		fmt.Printf("Current Capacity:    %s", capacity)
		fmt.Printf("Start Threshold:     %s", start_old)
		fmt.Printf("End Threshold:       %s", end_old)
		return
	}
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Must pass two arguments\n")
		os.Exit(1)
	}

	// check integer parsability
	new_start, err1 := strconv.Atoi(os.Args[1])
	new_end, err2 := strconv.Atoi(os.Args[2])
	if err1 != nil || err2 != nil {
		fmt.Println("Error converting Arg1 to int: ", err1 != nil)
		fmt.Println("Error converting Arg2 to int: ", err2 != nil)
		os.Exit(1)
	}

	// bounds check inputs 
	err3 := (new_start < new_end) 
	err4 := (0 <= new_start) && (new_start <= 100)
	err5 := (0 <= new_end) && (new_end <= 100)

	if !err3 || !err4 || !err5 {
		fmt.Println("Start < End:       ", err3)
		fmt.Println("0 <= Start <= 100: ", err4)
		fmt.Println("0 <= End <= 100:   ", err5)
		os.Exit(1)
	}

	// update kernel files
	start_res := update_kernel_param(startThresh, os.Args[1])
	end_res := update_kernel_param(endThresh, os.Args[2])

	if start_res != nil || end_res != nil {
		fmt.Println("error updating kernel files, ensure to run with sudo permissions")
		os.Exit(1)
	}

	fmt.Printf("Current Capacity:    %s", capacity)
	fmt.Printf("New Start Threshold: %s\n", os.Args[1])
	fmt.Printf("New End Threshold:   %s\n", os.Args[2])
}

func update_kernel_param(filepath string, val string) (error) {
	
	write_file, err1 := os.OpenFile(filepath, os.O_WRONLY, 0644)
	if err1 != nil {
		fmt.Printf("error opening %s\n", filepath)
		return errors.New("error opening file")
	}
	defer write_file.Close()

	_, err2 := write_file.WriteString(val)
	if err2 != nil {
		fmt.Printf("error writing to %s\n", filepath)
		return errors.New("error writing to file")
	}
	return nil
}

func read_current_levels () (string, string, string) {

	start_file_read, err1 := os.Open(startThresh)
	end_file_read, err2 := os.Open(endThresh)
	capacity_file_read, err3 := os.Open(capacityLevel)

	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Println("error opening files")
		os.Exit(1)
	}
	defer start_file_read.Close()
	defer end_file_read.Close()
	defer capacity_file_read.Close()

	init_start, err4 := io.ReadAll(start_file_read)
	init_end, err5 := io.ReadAll(end_file_read)
	cap_level, err6 := io.ReadAll(capacity_file_read)

	if err4 != nil || err5 != nil || err6 != nil {
		fmt.Println("error reading file contents")
		os.Exit(1)
	}
	return string(init_start), string(init_end), string(cap_level)
}