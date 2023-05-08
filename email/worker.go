package email

import "log"

type EmailJob struct {
	Type string
	Data KhoomiEmailData
}
type EmailWorker struct {
	Jobs chan EmailJob
	Quit chan bool
}

type EmailWorkerPool struct {
	Jobs    chan EmailJob
	Workers []EmailWorker
}

func KhoomiEmailWorkerPoolInstance(size int) *EmailWorkerPool {
	jobs := make(chan EmailJob, size)
	workers := make([]EmailWorker, size)

	for i := 0; i < size; i++ {
		workers[i] = EmailWorker{
			Jobs: jobs,
			Quit: make(chan bool),
		}
	}

	return &EmailWorkerPool{Jobs: jobs, Workers: workers}
}

func (pool *EmailWorkerPool) Start() {
	for id, worker := range pool.Workers {
		log.Printf("Email worker with %d started!\n", id)
		go worker.Start()
	}
}

func (pool *EmailWorkerPool) Stop() {
	for id, worker := range pool.Workers {
		log.Printf("Email worker with %d stopped!!\n", id)
		go worker.Stop()
	}
}

func (pool *EmailWorkerPool) Enqueue(job EmailJob) {
	pool.Jobs <- job
}

func (w *EmailWorker) Start() {
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

func (w *EmailWorker) Stop() {
	w.Quit <- true
}
