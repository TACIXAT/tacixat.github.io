Date = 2021-09-06T13:37:00-04:00
Published = true
[Meta]
Title = "A Better Parallelization Pattern for Babuk Ransomware"
Description = "Putting up because I talked some trash in my last post."
---

In my [previous post](https://tacix.at/posts/Babuk%20Source%20Code%20Leak%20-%20Golang%20Encryptor.html) analyzing the [Babuk Golang source code](https://gist.github.com/TACIXAT/92f04e033939136aa0171ff29a726e7a) I commented on the worker pattern. It would start a batch of threads but not kick off any more jobs until the slowest one completed. If you had 4 threads, you can imagine a case where 3 encrypt small files and then sit and do nothing while 1 finishes encrypting a very large file.

I've made a reproduction of theirs and then an alternative pattern using a workers and a job channel. To make a fair comparison, I keep a fixed set of jobs so both instances are processing the same data. The data processing is actually just sleeping for a random amount of seconds, but it will be the same random amount of seconds each time. At the top of our `main()` function we have some tunables and then we set up the jobs. 

```golang
numWorkers := 4
numJobs := 9
var wg sync.WaitGroup

rand.Seed(1337)
fixedJobs := make([]Job, numJobs)
for i := 0; i < numJobs; i++ {
	fixedJobs[i] = Job{
		Id: i,
		Seconds: time.Duration(rand.Intn(8)+2),
	}
}
```

The Babuk worker is simple, it just processes the job then marks itself as done in the `WaitGroup`. 

```golang
func worker(wg *sync.WaitGroup, id int, job Job) {
	defer wg.Done()

	log.Printf(
		"Worker %d: pretending to work for %d seconds for job %d", 
		id, job.Seconds, job.Id)
	time.Sleep(job.Seconds * time.Second)
	log.Printf("Worker %d: finished job %d", id, job.Id);
}
```

Back in `main()` they jobs are passed to a Go thread, we pass in the index of the jobs as the worker ID, and the job struct which consists of a job ID and a duration in seconds. We add 1 to the `activeWorkers` count. When the `activeWorkers` count reaches `numWorkers` we wait for the `WaitGroup` to finish, reset the count, then continue.

```golang
start := time.Now()
activeWorkers := 0
for i, j := range fixedJobs {
	go worker(&wg, i, j)
	wg.Add(1)
	activeWorkers += 1
	if activeWorkers == numWorkers {
		wg.Wait()
		activeWorkers = 0
	}
}
wg.Wait()
log.Printf("processed %d jobs in %s\n\n", numJobs, time.Since(start))
```

As stated before, this will cause jobs to happen in batches and be blocked by the slowest job. You can see this happening in the output below. In total, it took 26 seconds to process these random jobs.

```
2021/09/06 16:22:56 Worker 3: pretending to work for 5 seconds for job 3
2021/09/06 16:22:56 Worker 1: pretending to work for 8 seconds for job 1
2021/09/06 16:22:56 Worker 0: pretending to work for 4 seconds for job 0
2021/09/06 16:22:56 Worker 2: pretending to work for 9 seconds for job 2
2021/09/06 16:23:00 Worker 0: finished job 0
2021/09/06 16:23:01 Worker 3: finished job 3
2021/09/06 16:23:04 Worker 1: finished job 1
2021/09/06 16:23:05 Worker 2: finished job 2
2021/09/06 16:23:05 Worker 4: pretending to work for 6 seconds for job 4
2021/09/06 16:23:05 Worker 5: pretending to work for 6 seconds for job 5
2021/09/06 16:23:05 Worker 6: pretending to work for 8 seconds for job 6
2021/09/06 16:23:05 Worker 7: pretending to work for 3 seconds for job 7
2021/09/06 16:23:08 Worker 7: finished job 7
2021/09/06 16:23:11 Worker 5: finished job 5
2021/09/06 16:23:11 Worker 4: finished job 4
2021/09/06 16:23:13 Worker 6: finished job 6
2021/09/06 16:23:13 Worker 8: pretending to work for 9 seconds for job 8
2021/09/06 16:23:22 Worker 8: finished job 8
2021/09/06 16:23:22 processed 9 jobs in 26.0143759s
```

## Revised Worker

For our code we will start `numWorkers` of Go threads. Then we will place the jobs into a channel. A channel is a bit like a thread safe queue. Workers will pull jobs out of the channel and process them, and then grab the next. This allows all workers to be fully occupied. We use a job ID of `-1` as a sentinel value, letting the workers know that they are done.

Our `main()` method will start the workers, queue the jobs, add a `numWorkers` number of `-1`s to the queue, then wait on the `WaitGroup`.

```golang
start = time.Now()
jobs := make(chan Job)
for i := 0; i < numWorkers; i++ {
	go queueWorker(&wg, i, jobs)
	wg.Add(1)
}

for _, j := range fixedJobs {
	jobs <- j
}

for i := 0; i < numWorkers; i++ {
	jobs <- Job{
		Id: -1,
	}
}

wg.Wait()
log.Printf("processed %d jobs in %s", numJobs, time.Since(start))
```

The worker is a little bit more complex. It has an infinite loop, grabs a job out of the channel, if it is a negative `job.Id` it will exit, if not it does the same processing as the other worker.

```golang
func queueWorker(wg *sync.WaitGroup, id int, jobs chan Job) {
	defer wg.Done()

	for {
		job := <-jobs

		if job.Id < 0 {
			return
		}

		log.Printf(
			"Worker %d: pretending to work for %d seconds for job %d", 
			id, job.Seconds, job.Id)
		time.Sleep(job.Seconds * time.Second)
		log.Printf("Worker %d: finished job %d", id, job.Id);
	}
}
```

Now, when a worker finishes their job, they can immediately move on to the next job. There is no blocking on the slowest of the group. Now, it processes all jobs in 19 seconds!

```
2021/09/06 16:23:22 Worker 0: pretending to work for 4 seconds for job 0
2021/09/06 16:23:22 Worker 2: pretending to work for 5 seconds for job 3
2021/09/06 16:23:22 Worker 3: pretending to work for 8 seconds for job 1
2021/09/06 16:23:22 Worker 1: pretending to work for 9 seconds for job 2
2021/09/06 16:23:26 Worker 0: finished job 0
2021/09/06 16:23:26 Worker 0: pretending to work for 6 seconds for job 4
2021/09/06 16:23:27 Worker 2: finished job 3
2021/09/06 16:23:27 Worker 2: pretending to work for 6 seconds for job 5
2021/09/06 16:23:30 Worker 3: finished job 1
2021/09/06 16:23:30 Worker 3: pretending to work for 8 seconds for job 6
2021/09/06 16:23:31 Worker 1: finished job 2
2021/09/06 16:23:31 Worker 1: pretending to work for 3 seconds for job 7
2021/09/06 16:23:32 Worker 0: finished job 4
2021/09/06 16:23:32 Worker 0: pretending to work for 9 seconds for job 8
2021/09/06 16:23:33 Worker 2: finished job 5
2021/09/06 16:23:34 Worker 1: finished job 7
2021/09/06 16:23:38 Worker 3: finished job 6
2021/09/06 16:23:41 Worker 0: finished job 8
2021/09/06 16:23:41 processed 9 jobs in 19.0381034s
```

There could always be a reason for the pattern they chose, perhaps to not saturate disk or cores in some way. That said, if you are looking to utilize cores to the fullest this pattern is much more efficient.