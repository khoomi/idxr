package email

import "log"

type Job struct {
	Type string
	Data KhoomiEmailData
}
type Worker struct {
	Jobs chan Job
	Quit chan bool
}

type WorkerPool struct {
	Jobs    chan Job
	Workers []Worker
}

func KhoomiEmailWorkerPoolInstance(size int) *WorkerPool {
	jobs := make(chan Job, size)
	workers := make([]Worker, size)

	for i := 0; i < size; i++ {
		workers[i] = Worker{
			Jobs: jobs,
			Quit: make(chan bool),
		}
	}

	return &WorkerPool{Jobs: jobs, Workers: workers}
}

func (pool *WorkerPool) Start() {
	for id, worker := range pool.Workers {
		log.Printf("Email worker with %d started!\n", id)
		go worker.Start()
	}
}

func (pool *WorkerPool) Stop() {
	for id, worker := range pool.Workers {
		log.Printf("Email worker with %d stopped!!\n", id)
		go worker.Stop()
	}
}

func (pool *WorkerPool) Enqueue(job Job) {
	pool.Jobs <- job
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.Jobs:
				// Send the email based on the job type and data
				switch job.Type {
				case "verify":
					log.Printf("KhoomiEmail: Sent email verification mail to user %s, - ip: %s", job.Data.Email, job.Data.IP)
					SendVerifyEmailNotification(job.Data)
				case "welcome":
					log.Printf("KhoomiEmail: Sent welcome email to a new registered user %s, - ip: %s", job.Data.Email, job.Data.IP)
					SendWelcomeEmail(job.Data)
				case "ipaddr":
					log.Printf("KhoomiEmail: User %s just logged in successfully from a new IP adrress  - ip: %s", job.Data.Email, job.Data.IP)
					SendNewIpLoginNotification(job.Data)
				case "password_reset":
					log.Printf("KhoomiEmail: User with email %s, requested to for a password reset from ip: %s", job.Data.Email, job.Data.IP)
					SendPasswordResetEmail(job.Data)
				case "password_reset_success":
					log.Printf("KhoomiEmail: Sending user password email reset successfully to %s - ip: %s", job.Data.Email, job.Data.IP)
					SendPasswordResetSuccessfulEmail(job.Data)
				}
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.Quit <- true
}
