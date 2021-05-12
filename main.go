package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Assessment tracks the content and results of the test.
type Assessment struct {
	Questions      []Question    //slice of Question stuct
	TotalCorrect   int           //Number of Questions answered correctly
	TotalIncorrect int           //number of Questions answered incorrectly[]
	TotalQuestions int           //Total number of Questions in Assessment
	FilePath       string        //Filepath to file contaning questions
	Shuffle        bool          //Should the questions be randomized / shuffled
	TimeLimit      time.Duration //The amount of time the user has to complete the test
	TimeStart      time.Time     //Start time for the Assessment
	Name           string        //Name of the user taking the Quiz
}

// ParseCmdLnArgs Reads the params from the commandline and sets
// values on the Assessment struct.
// It is called from LoadQuestions.
// If the user adds -help, -h, or help arguments to the command then the program will show help and then exit.
func (a *Assessment) ParseCmdLnArgs() {

	// Setup the help flag so that the message is displayed when the user asks for help
	var flaghelp string
	flag.StringVar(&flaghelp, "help", "", "Print this help text")
	flag.StringVar(&flaghelp, "h", "", "Print this help text")

	// Flag duration requires a time.Duration object so we set it here
	var DefaultTimeLimit time.Duration = time.Second * 30 //30 seconds

	//These variables are the commandline flags which are parsed by the flags module
	flagfilepath := flag.String("filepath", "problems.csv", "A CSV file containing quiz questions")
	flagshuffle := flag.Bool("shuffle", false, "When set to True, the quiz questions are shuffled. (default \"false\")")
	flagtotalquestions := flag.Int("totalquestions", 0, "Number of questions in the test.\nIf no count is provided then all questions in the file will be used.")
	flagtimelimit := flag.Duration("timelimit", DefaultTimeLimit, "Time limit for the test")

	flag.Parse()

	// After the flags are parsed, we store the data in the Assessment struct
	a.FilePath = *flagfilepath
	a.Shuffle = *flagshuffle
	a.TotalQuestions = *flagtotalquestions
	a.TimeLimit = *flagtimelimit

	// if the user passed -help, -h, or help to the command then show help and exit
	for _, v := range os.Args {
		v = strings.Trim(v, " ")
		if v == "-help" || v == "-h" || v == "help" {
			fmt.Println("------------------------")
			fmt.Println("quiz - play a quiz game")
			fmt.Println("** syntax -var=Value **")
			flag.PrintDefaults()
			fmt.Println("------------------------")
			os.Exit(0) //show the help and exit the program
		}
	}
}

// ShuffleQuestions will shuffle the questions in the Questions slice of the Assessment struct.
// This function is called from LoadQuestions.
func (a *Assessment) ShuffleQuestions() {
	if len(a.Questions) == 0 || !a.Shuffle { //if there are no questions don't do anything
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a.Questions), func(i, j int) { a.Questions[i], a.Questions[j] = a.Questions[j], a.Questions[i] })
}

// LoadQuestions loads a csv file containing questions and answers.
// it returns an error if loading fails.
func (a *Assessment) LoadQuestions() (err error) {

	a.ParseCmdLnArgs()

	file, err := os.Open(a.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	if a.TotalQuestions > len(records) || a.TotalQuestions == 0 {
		a.TotalQuestions = len(records)
	}

	// Load each Question Struct into the Questions Slice in the Assessment Struct
	for i := 0; i < a.TotalQuestions; i++ {
		v := records[i]
		question := Question{QText: v[0], Answer: v[1]}
		a.Questions = append(a.Questions, question)
	}

	// Shuffle the questions if needed
	a.ShuffleQuestions()

	return nil
}

func (a *Assessment) GreetUser() (err error) {
	fmt.Println("Welcome to the Quiz Game")
	fmt.Printf("Please enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	a.Name, err = reader.ReadString('\n')

	a.Name = strings.TrimSpace(a.Name)

	if err != nil {
		fmt.Println("Error occurred:", err)
		return err
	}
	return nil
}

// StartTest administers the test by looping through the questions in the Questions slice
// and setting the properties on the Assessment struct.
// it also runs the timer for the test.
func (a *Assessment) StartTest() (err error) {

	err = a.GreetUser()
	if err != nil {
		fmt.Println("Error occurred:", err)
		return err
	}

	fmt.Printf("You have %s to finish the test. There are %v questions in the test.\nPress ENTER to start the test", a.TimeLimit, a.TotalQuestions)
	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error occurred:", err)
		os.Exit(1)
	}
	a.TimeStart = time.Now()
	timer := time.AfterFunc(a.TimeLimit, func() {
		fmt.Println("")
		fmt.Printf("Time's Up %s!\n", a.Name)
		a.ShowScore()
		os.Exit(0)
	})
	defer timer.Stop()

	for i := 0; i < len(a.Questions); i++ {
		err := a.Questions[i].AskQuestion(i + 1)

		if err != nil {
			return err
		}
		if a.Questions[i].Correct {
			a.TotalCorrect++
		} else {
			a.TotalIncorrect++
		}
	}
	a.ShowScore()

	return nil
}

// ShowScore prints out the results of the test.
func (a *Assessment) ShowScore() {

	// if the user answered all the questions then tell them
	// how much time they took to answer the questions and how
	// much time was left on the clock
	if a.TotalCorrect+a.TotalIncorrect == a.TotalQuestions {
		Now := time.Now()
		TestTime := Now.Sub(a.TimeStart)
		TimeLeft := a.TimeLimit.Seconds() - TestTime.Seconds()

		fmt.Printf("You answered all %v questions in %.2f seconds.\nThere were %.2f seconds remaining on the clock.\n",
			a.TotalQuestions, TestTime.Seconds(), TimeLeft)
	} else {
		fmt.Printf("You answered %v questions out of a total of %v questions in %.2f seconds.\n",
			a.TotalCorrect+a.TotalIncorrect, a.TotalQuestions, a.TimeLimit.Seconds())
	}
	fmt.Printf("You got %v questions right and %v questions wrong.\n", a.TotalCorrect, a.TotalIncorrect)
	fmt.Printf("Your score is %.2f%% %s! \n", float32(a.TotalCorrect)/float32(a.TotalQuestions)*100, a.Name)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Question", "Answer", "User Answer", "Correct"})

	for i, v := range a.Questions {
		table.Append([]string{strconv.FormatInt(int64(i+1), 10), v.QText, v.Answer, v.UserAnswer, strconv.FormatBool(v.Correct)})
	}

	table.Render() // Send output
}

// Question struct stores the fields for each question in the assessment.
type Question struct {
	QText      string //Question text
	Answer     string //Correct Answer for Question
	UserAnswer string //Answer the user Provided
	Correct    bool   //Whether the user got the answer right or not
}

// AskQuestion delivers a question and tracks the user's response in the
// Question struct.  The qnum variable tracks the number for the question.
func (q *Question) AskQuestion(qnum int) (err error) {
	fmt.Printf("%v. %s = ", qnum, q.QText)
	reader := bufio.NewReader(os.Stdin)
	q.UserAnswer, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error occurred:", err)
		return err
	}
	q.UserAnswer = strings.TrimSpace(q.UserAnswer)

	if q.UserAnswer == q.Answer { // Answer is correct
		q.Correct = true
	}

	return nil
}

func main() {
	var test Assessment

	err := test.LoadQuestions()
	if err != nil {
		log.Panic("Unable to load questions.  The following error occured: ", err)
	}

	err = test.StartTest()
	if err != nil {
		log.Panic("Unable to aAdminister test.  The following error occured: ", err)
	}
}
