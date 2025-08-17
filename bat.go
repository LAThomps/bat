package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const batFiles string = "/sys/class/power_supply/BAT0/"
const startThresh string = batFiles + "charge_control_start_threshold"
const endThresh string = batFiles + "charge_control_end_threshold"
const capacityLevel string = batFiles + "capacity"
const statusLevel string = batFiles + "status"

func main() {

	start_old, end_old, capacity, status := read_current_levels()
	
	if len(os.Args) == 1 {
		charge_time, err0 := calc_charge_time(capacity, end_old)
		time_to_20, err1 := calc_time_to_perc(capacity, 20)
		time_to_0, err2 := calc_time_to_perc(capacity, 0)
		if err0 != nil || err1 != nil || err2 != nil {
			fmt.Println("Error calculating times")
			fmt.Printf("Charge Time:   %s", err0)
			fmt.Printf("Time to 20:    %s", err1)
			fmt.Printf("Time to 0:     %s", err2)
			os.Exit(1)
		}

		fmt.Printf("Current Capacity:    %s\n", capacity)
		fmt.Printf("Start Threshold:     %s\n", start_old)
		fmt.Printf("End Threshold:       %s\n", end_old)
		fmt.Printf("Status:              %s\n", status)
		fmt.Printf("Time to End Charge:  %s mins\n", charge_time)
		fmt.Printf("Time to 20:          %s\n", time_to_20)
		fmt.Printf("Time to 0:           %s\n", time_to_0)
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
		fmt.Println("Error converting Arg1 to int: ", err1)
		fmt.Println("Error converting Arg2 to int: ", err2)
		os.Exit(1)
	}

	// bounds check inputs
	err3 := (new_start < new_end)
	err4 := (0 <= new_start) && (new_start <= 100)
	err5 := (0 <= new_end) && (new_end <= 100)

	if !err3 || !err4 || !err5 {
		fmt.Println("Invalid Input Arguments:")
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

	charge_time, err6 := calc_charge_time(capacity, os.Args[2])
	if err6 != nil {
		fmt.Println("Error calculating charge time")
		os.Exit(1)
	}

	// final result
	fmt.Printf("Current Capacity:    %s\n", capacity)
	fmt.Printf("New Start Threshold: %s\n", os.Args[1])
	fmt.Printf("New End Threshold:   %s\n", os.Args[2])
	fmt.Printf("Status:              %s\n", status)
	fmt.Printf("Time to End Charge:  %s\n", charge_time)
}

func calc_time_to_perc(capacity_str string, end_perc int) (string, error) {
	
	capacity, err := strconv.Atoi(capacity_str)
	if err != nil {
		return "NA", err
	}

	hours_left := ((capacity - end_perc) / 14) // normal usage 14% per hour
	mins_left := (((float32(capacity) - float32(end_perc)) / 14) -
		float32(hours_left)) * 60
	result_msg := fmt.Sprintf("%d hour(s) %.0f mins", hours_left, mins_left)
	return result_msg, nil
}

func calc_charge_time(capacity_str string, end_thresh_str string) (string, error) {

	capacity, err1 := strconv.Atoi(capacity_str)
	end_thresh, err2 := strconv.Atoi(end_thresh_str)
	if err1 != nil || err2 != nil {
		fmt.Println("Could not convert calcs to int:")
		fmt.Println("capacity_str:   ", err1)
		fmt.Println("end_thresh_str: ", err2)
		return "NA", errors.New("could not calculate charge time")
	}

	if capacity > end_thresh {
		return "0", nil
	} 

	perc_to_charge := (float32(end_thresh) - float32(capacity)) / 100
	charge_time := perc_to_charge * (52.5 / (65 * 0.85))
	charge_time_mins := strconv.Itoa(int(charge_time * 60))
	return charge_time_mins, nil
}

func update_kernel_param(filepath string, val string) error {

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

func read_current_levels() (string, string, string, string) {

	init_start, err0 := read_file(startThresh)
	init_end, err1 := read_file(endThresh)
	cap_level, err2 := read_file(capacityLevel)
	status, err3 := read_file(statusLevel)

	if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
		fmt.Println("Error opening kernel files")
		fmt.Printf("Start Threshold File:  %s", err0)
		fmt.Printf("End Threshold File:    %s", err1)
		fmt.Printf("Capacity File:         %s", err2)
		fmt.Printf("Status:                %s", err3)
		os.Exit(1)
	}

	return init_start, init_end, cap_level, status
}

func read_file(filepath string) (string, error) {

	file_read, err1 := os.Open(filepath)
	if err1 != nil {
		return "NA", errors.New("error opening file")
	}
	defer file_read.Close()

	level_read, err2 := io.ReadAll(file_read)
	if err2 != nil {
		return "NA", errors.New("error reading file")
	}

	level_actual := strings.TrimSpace(string(level_read))
	return level_actual, nil
}
