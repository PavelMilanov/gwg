package tc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PavelMilanov/go-wg-manager/server"
)

var (
	ratePattern          = regexp.MustCompile(`^[1-9][0-9]*(?:Kbit|Mbit|Gbit|kbit|mbit|gbit|bit)$`)
	identifierPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	interfaceNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.+=-]{0,14}$`)
)

func UpService(speed, fullSpeed string) error {
	if speed == "" {
		speed = fullSpeed
	}
	if err := validateRate("speed", speed); err != nil {
		return err
	}
	if err := validateRate("full speed", fullSpeed); err != nil {
		return err
	}
	state, stateErr := serviceState()
	if state == "enabled" {
		return fmt.Errorf("tc service is already enabled")
	}
	if stateErr != nil && state != "disabled" && state != "not-found" {
		return stateErr
	}
	if err := ensureTcDir(); err != nil {
		return err
	}
	classes, err := readClassFile()
	if err != nil {
		return err
	}
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	wgServer, err := server.ReadServerConfigFile()
	if err != nil {
		return err
	}
	tc := TcConfig{Intf: wgServer.Alias, Speed: speed, FullSpeed: fullSpeed, Classes: classes, Filters: filters}
	if err := validateConfig(tc); err != nil {
		return err
	}
	if err := tc.config(); err != nil {
		return err
	}
	if err := tc.generate(); err != nil {
		return err
	}
	if err := tc.createService(); err != nil {
		return err
	}
	fmt.Println("Gwg tc service started")
	return nil
}

func DownService() error {
	tc, err := readTcFile()
	if err != nil {
		return err
	}
	if err := errors.Join(tc.removeService(), tc.down()); err != nil {
		return err
	}
	fmt.Println("Gwg tc service down")
	return nil
}

func RestartService() error {
	tc, err := readTcFile()
	if err != nil {
		return err
	}
	classes, err := readClassFile()
	if err != nil {
		return err
	}
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	tc.Classes = classes
	tc.Filters = filters
	if err := validateConfig(tc); err != nil {
		return err
	}
	if err := tc.config(); err != nil {
		return err
	}
	if err := tc.generate(); err != nil {
		return err
	}
	if err := tc.down(); err != nil {
		return err
	}
	if err := tc.start(); err != nil {
		return err
	}
	fmt.Println("Gwg tc service restarted")
	return nil
}

func ShowService() error {
	state, err := serviceState()
	if err != nil || state != "enabled" {
		return fmt.Errorf("tc service is not enabled")
	}
	tc, err := readTcFile()
	if err != nil {
		return err
	}
	fmt.Printf("Gwg tc service:\n\tFullSpeed: %s\n\tSpeed: %s\n\tClasses: %v\n\tFilters: %v\n", tc.FullSpeed, tc.Speed, tc.Classes, tc.Filters)
	return nil
}

func serviceState() (string, error) {
	out, err := commandRunner.Output("sudo", "systemctl", "is-enabled", tcServiceFile)
	return strings.TrimSpace(string(out)), err
}

func AddBandwidth(description, minSpeed, ceilSpeed string) error {
	if err := validateIdentifier("description", description); err != nil {
		return err
	}
	if err := validateRate("minimum speed", minSpeed); err != nil {
		return err
	}
	if err := validateRate("ceiling speed", ceilSpeed); err != nil {
		return err
	}
	if err := ensureTcDir(); err != nil {
		return err
	}
	configs, err := readClassFile()
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.Description == description {
			return fmt.Errorf("bandwidth %q already exists", description)
		}
	}
	classID := nextClassID(configs)
	config := TcClass{Class: classID, Description: description, MinSpeed: minSpeed, CeilSpeed: ceilSpeed}
	configs = append(configs, config)
	if err := writeJSONFile(classFile, configs); err != nil {
		return err
	}
	printClass(config)
	fmt.Println("Added successfully")
	return nil
}

func nextClassID(classes []TcClass) string {
	used := make(map[int]struct{}, len(classes))
	for _, class := range classes {
		id, err := strconv.Atoi(class.Class)
		if err == nil {
			used[id] = struct{}{}
		}
	}
	for id := 2; ; id++ {
		if _, exists := used[id]; !exists {
			return strconv.Itoa(id)
		}
	}
}

func RemoveBandwidth(classID string) error {
	if _, err := strconv.ParseUint(classID, 10, 16); err != nil || classID == "1" {
		return fmt.Errorf("invalid class id %q", classID)
	}
	configs, err := readClassFile()
	if err != nil {
		return err
	}
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	for _, filter := range filters {
		if filter.Class == classID {
			return fmt.Errorf("class %s is used by filter %q", classID, filter.Description)
		}
	}
	index := -1
	for i, config := range configs {
		if config.Class == classID {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("class %s not found", classID)
	}
	removed := configs[index]
	configs = append(configs[:index], configs[index+1:]...)
	if err := writeJSONFile(classFile, configs); err != nil {
		return err
	}
	printClass(removed)
	fmt.Println("Removed successfully")
	return nil
}

func ShowBandwidth() error {
	configs, err := readClassFile()
	if err != nil {
		return err
	}
	sort.Slice(configs, func(i, j int) bool {
		left, _ := strconv.Atoi(configs[i].Class)
		right, _ := strconv.Atoi(configs[j].Class)
		return left < right
	})
	for _, config := range configs {
		printClass(config)
		fmt.Println()
	}
	return nil
}

func printClass(config TcClass) {
	fmt.Printf("class: %s\n\tdescription: %s;\n\tmin-rate: %s;\n\tceil-rate: %s;\n", config.Class, config.Description, config.MinSpeed, config.CeilSpeed)
}

func AddFilter(description, userName, classID string) error {
	if err := validateIdentifier("description", description); err != nil {
		return err
	}
	classes, err := readClassFile()
	if err != nil {
		return err
	}
	classFound := false
	for _, item := range classes {
		if item.Class == classID {
			classFound = true
			break
		}
	}
	if !classFound {
		return fmt.Errorf("class %q not found", classID)
	}
	users, err := server.ReadClientConfigFiles()
	if err != nil {
		return err
	}
	var userIP string
	for _, item := range users {
		if item.Name == userName {
			userIP = item.ClientLocalAddress
			break
		}
	}
	if userIP == "" {
		return fmt.Errorf("user %q not found", userName)
	}
	if _, err := netip.ParsePrefix(userIP); err != nil {
		return fmt.Errorf("user %q has invalid IP prefix: %w", userName, err)
	}
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	for _, filter := range filters {
		if filter.Description == description {
			return fmt.Errorf("filter %q already exists", description)
		}
	}
	filter := TcFilter{Description: description, UserIp: userIP, Class: classID}
	filters = append(filters, filter)
	if err := writeJSONFile(filterFile, filters); err != nil {
		return err
	}
	printFilter(filter)
	fmt.Println("Added successfully")
	return nil
}

func RemoveFilter(description string) error {
	if err := validateIdentifier("filter", description); err != nil {
		return err
	}
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	index := -1
	for i, filter := range filters {
		if filter.Description == description {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("filter %q not found", description)
	}
	removed := filters[index]
	filters = append(filters[:index], filters[index+1:]...)
	if err := writeJSONFile(filterFile, filters); err != nil {
		return err
	}
	printFilter(removed)
	fmt.Println("Removed successfully")
	return nil
}

func ShowFilter() error {
	filters, err := readFilterFile()
	if err != nil {
		return err
	}
	for _, filter := range filters {
		printFilter(filter)
		fmt.Println()
	}
	return nil
}

func printFilter(filter TcFilter) {
	fmt.Printf("filter: %s\n\tuser: %s;\n\tclass: %s;\n", filter.Description, filter.UserIp, filter.Class)
}

func readClassFile() ([]TcClass, error) {
	var classes []TcClass
	err := readOptionalJSON(classFile, &classes)
	return classes, err
}

func readFilterFile() ([]TcFilter, error) {
	var filters []TcFilter
	err := readOptionalJSON(filterFile, &filters)
	return filters, err
}

func readTcFile() (TcConfig, error) {
	var config TcConfig
	path := filepath.Join(tcDir, tcFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return TcConfig{}, fmt.Errorf("read tc configuration: %w", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return TcConfig{}, fmt.Errorf("decode tc configuration: %w", err)
	}
	if err := validateConfig(config); err != nil {
		return TcConfig{}, fmt.Errorf("invalid tc configuration: %w", err)
	}
	return config, nil
}

func readOptionalJSON(name string, target any) error {
	path := filepath.Join(tcDir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func validateConfig(config TcConfig) error {
	if err := validateInterfaceName(config.Intf); err != nil {
		return err
	}
	if err := validateRate("speed", config.Speed); err != nil {
		return err
	}
	if err := validateRate("full speed", config.FullSpeed); err != nil {
		return err
	}
	classes := make(map[string]struct{}, len(config.Classes))
	for _, class := range config.Classes {
		id, err := strconv.ParseUint(class.Class, 10, 16)
		if err != nil || id < 2 {
			return fmt.Errorf("invalid class id %q", class.Class)
		}
		if _, exists := classes[class.Class]; exists {
			return fmt.Errorf("duplicate class id %q", class.Class)
		}
		classes[class.Class] = struct{}{}
		if err := validateIdentifier("class description", class.Description); err != nil {
			return err
		}
		if err := validateRate("minimum speed", class.MinSpeed); err != nil {
			return err
		}
		if err := validateRate("ceiling speed", class.CeilSpeed); err != nil {
			return err
		}
	}
	filterNames := make(map[string]struct{}, len(config.Filters))
	for _, filter := range config.Filters {
		if err := validateIdentifier("filter description", filter.Description); err != nil {
			return err
		}
		if _, exists := filterNames[filter.Description]; exists {
			return fmt.Errorf("duplicate filter %q", filter.Description)
		}
		filterNames[filter.Description] = struct{}{}
		prefix, err := netip.ParsePrefix(filter.UserIp)
		if err != nil || !prefix.Addr().Is4() || prefix.Bits() != 32 {
			return fmt.Errorf("invalid filter IP %q", filter.UserIp)
		}
		if _, exists := classes[filter.Class]; !exists {
			return fmt.Errorf("filter %q references missing class %q", filter.Description, filter.Class)
		}
	}
	return nil
}

func validateRate(name, value string) error {
	if !ratePattern.MatchString(value) {
		return fmt.Errorf("invalid %s %q (example: 50Mbit)", name, value)
	}
	return nil
}

func validateIdentifier(name, value string) error {
	if !identifierPattern.MatchString(value) || value == "." || value == ".." {
		return fmt.Errorf("invalid %s %q", name, value)
	}
	return nil
}

func validateInterfaceName(value string) error {
	if !interfaceNamePattern.MatchString(value) || value == "." || value == ".." {
		return fmt.Errorf("invalid WireGuard interface name %q", value)
	}
	return nil
}
