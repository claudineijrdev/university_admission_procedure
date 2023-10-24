package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type SortApplicants interface {
	sortApplicants() []Applicant
}

type Applicant struct {
	name              string
	grades            map[string]float64
	departmentOptions []string
	accepted          bool
}

func NewApplicant(name string, examGrades []float64, departmentOption []string) Applicant {
	grades := make(map[string]float64, 0)
	grades["physics"] = examGrades[0]
	grades["chemistry"] = examGrades[1]
	grades["math"] = examGrades[2]
	grades["computer science"] = examGrades[3]
	grades["special"] = examGrades[4]

	return Applicant{
		name:              name,
		grades:            grades,
		departmentOptions: departmentOption,
		accepted:          false,
	}
}

func (a *Applicant) calculateMean(exams []string) float64 {
	grade := 0.0
	for _, exam := range exams {
		grade += a.grades[exam]
	}
	mean := grade / float64(len(exams))
	if a.grades["special"] > mean {
		return a.grades["special"]
	}
	return mean
}

type Department struct {
	name       string
	exams      []string
	limit      int
	applicants []*Applicant
	accepted   []*Applicant
}

func NewDepartment(name string, limit int, exams []string) Department {
	return Department{
		name:  name,
		exams: exams,
		limit: limit,
	}
}

func (d *Department) addApplicant(applicant *Applicant) {
	d.applicants = append(d.applicants, applicant)
}

func (d *Department) sortApplicants() {
	sort.Slice(d.applicants, func(i, j int) bool {
		if d.applicants[i].calculateMean(d.exams) != d.applicants[j].calculateMean(d.exams) {
			return d.applicants[i].calculateMean(d.exams) > d.applicants[j].calculateMean(d.exams)
		}
		return d.applicants[i].name < d.applicants[j].name
	})
}

func (d *Department) sortAccepted() {
	sort.Slice(d.accepted, func(i, j int) bool {
		if d.accepted[i].calculateMean(d.exams) != d.accepted[j].calculateMean(d.exams) {
			return d.accepted[i].calculateMean(d.exams) > d.accepted[j].calculateMean(d.exams)
		}
		return d.accepted[i].name < d.accepted[j].name
	})
}

type University struct {
	departments    []*Department
	mapDepartments map[string]int
	applicants     []*Applicant
	calls          int
}

func NewUniversity(departments []*Department, applicants []*Applicant, calls int) University {
	mapDepartments := make(map[string]int, len(departments))
	for index, department := range departments {
		mapDepartments[department.name] = index
	}

	return University{
		departments:    departments,
		applicants:     applicants,
		mapDepartments: mapDepartments,
		calls:          calls,
	}
}

func (u *University) segmentsApplicants() {
	for _, applicant := range u.applicants {
		for _, option := range applicant.departmentOptions {
			departmentIndex := u.mapDepartments[option]
			u.departments[departmentIndex].addApplicant(applicant)
		}
	}
}

func (u *University) sortApplicants() {
	for index := range u.departments {
		u.departments[index].sortApplicants()
	}
}

func (u *University) sortDepartments() {
	for index := range u.departments {
		u.departments[index].sortAccepted()
	}
	sort.Slice(u.departments, func(i, j int) bool {
		return u.departments[i].name < u.departments[j].name
	})
}

func (u *University) selectCandidates() {
	for i := 0; i < u.calls; i++ {
		for _, department := range u.departments {
			for _, applicant := range department.applicants {
				if len(department.accepted) >= department.limit {
					break
				}
				if !applicant.accepted && applicant.departmentOptions[i] == department.name {
					department.accepted = append(department.accepted, applicant)
					applicant.accepted = true
				}
			}
		}
	}
}

func scanApplicants(file *os.File) (applicants []*Applicant) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		tuple := strings.Split(scanner.Text(), " ")
		name := fmt.Sprintf("%s %s", tuple[0], tuple[1])
		examGrades := make([]float64, 0)
		for i := 2; i < 7; i++ {
			grade := strings.ReplaceAll(tuple[i], ",", ".")
			gradeFloat, _ := strconv.ParseFloat(grade, 64)
			examGrades = append(examGrades, gradeFloat)
		}
		departmentOptions := tuple[7:]
		applicant := NewApplicant(name, examGrades, departmentOptions)
		applicants = append(applicants, &applicant)
	}
	return applicants
}

func loadApplicantList(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return file
}

func printResults(university *University) {
	for _, department := range university.departments {
		fmt.Println(department.name)
		for _, applicant := range department.accepted {
			fmt.Printf("%s %.1f\n", applicant.name, applicant.calculateMean(department.exams))
		}
		fmt.Println()
	}
}

func saveResults(university *University) {
	for _, department := range university.departments {
		file, err := os.Create(strings.ToLower(department.name) + ".txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		for _, applicant := range department.accepted {
			fmt.Fprintf(file, "%s %.1f\n", applicant.name, applicant.calculateMean(department.exams))
		}
		file.WriteString("\n")
	}
}

func main() {
	path := "applicants.txt"
	var limit int
	fmt.Scan(&limit)

	physics := NewDepartment("Physics", limit, []string{"physics", "math"})
	chemistry := NewDepartment("Chemistry", limit, []string{"chemistry"})
	math := NewDepartment("Mathematics", limit, []string{"math"})
	engineering := NewDepartment("Engineering", limit, []string{"computer science", "math"})
	biotech := NewDepartment("Biotech", limit, []string{"chemistry", "physics"})

	applicants := scanApplicants(loadApplicantList(path))
	university := NewUniversity([]*Department{
		&physics,
		&chemistry,
		&math,
		&engineering,
		&biotech,
	}, applicants, 3)

	university.segmentsApplicants()
	university.sortApplicants()
	university.selectCandidates()
	university.sortDepartments()

	printResults(&university)
	saveResults(&university)
}
