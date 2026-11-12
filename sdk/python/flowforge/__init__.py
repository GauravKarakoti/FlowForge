import logging
import time
from typing import Any, Callable, Dict, List, Optional
import requests

logger = logging.getLogger("flowforge")


class Client:
    """
    Client for interacting with the FlowForge control plane.
    """

    def __init__(self, host: str = "http://localhost:8080", token: Optional[str] = None):
        self.host = host.rstrip("/")
        self.token = token
        # Initialize HTTP session
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})
        if self.token:
            self.session.headers.update({"Authorization": f"Bearer {self.token}"})

    def enqueue(
        self,
        task_name: str,
        args: Optional[List[Any]] = None,
        kwargs: Optional[Dict[str, Any]] = None,
        queue: str = "default",
        retry_limit: int = 3,
    ) -> Dict[str, Any]:
        """
        Submits a job to the FlowForge control plane.
        """
        payload = {
            "task_name": task_name,
            "args": args if args is not None else [],
            "kwargs": kwargs if kwargs is not None else {},
            "queue": queue,
            "retry_limit": retry_limit,
        }
        url = f"{self.host}/api/jobs/enqueue"
        response = self.session.post(url, json=payload)
        response.raise_for_status()
        try:
            return response.json()
        except ValueError:
            return {"status": "enqueued", "status_code": response.status_code}


class Worker:
    """
    Worker node for polling and executing tasks assigned to specific queues.
    """

    def __init__(
        self,
        queue: str = "default",
        host: str = "http://localhost:8080",
        token: Optional[str] = None,
        poll_interval: float = 1.0,
    ):
        self.queue = queue
        self.host = host.rstrip("/")
        self.token = token
        self.poll_interval = poll_interval
        self.registry: Dict[str, Callable] = {}
        self._running = False

        # Setup connection to broker or control plane
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})
        if self.token:
            self.session.headers.update({"Authorization": f"Bearer {self.token}"})

    def register_task(self, name: str, func: Callable) -> None:
        """
        Registers a local python function to handle a specific task name.
        """
        self.registry[name] = func
        logger.info(f"Registered task '{name}'")

    def report_status(
        self,
        job_id: str,
        status: str,
        result: Optional[Any] = None,
        error: Optional[str] = None,
    ) -> None:
        """
        Reports execution state of a job back to the control plane.
        """
        url = f"{self.host}/api/jobs/{job_id}/status"
        payload = {"status": status, "result": result, "error": error}
        try:
            res = self.session.post(url, json=payload, timeout=5)
            res.raise_for_status()
        except Exception as e:
            logger.error(f"Failed to report status for job {job_id}: {e}")

    def poll_job(self) -> Optional[Dict[str, Any]]:
        """
        Polls the control plane for the next available job in the worker's queue.
        """
        url = f"{self.host}/api/jobs/poll"
        try:
            res = self.session.get(url, params={"queue": self.queue}, timeout=10)
            if res.status_code == 200:
                data = res.json()
                if data and "job" in data:
                    return data["job"]
                return data if data else None
            elif res.status_code in (204, 404):
                return None
            else:
                logger.warning(f"Unexpected status code {res.status_code} during poll: {res.text}")
                return None
        except requests.RequestException as e:
            logger.debug(f"Polling error: {e}")
            return None

    def execute_job(self, job: Dict[str, Any]) -> None:
        """
        Executes a fetched job and reports its status back to the control plane.
        """
        job_id = job.get("id") or job.get("job_id")
        task_name = job.get("task_name")
        args = job.get("args", [])
        kwargs = job.get("kwargs", {})

        if task_name not in self.registry:
            err = f"No handler registered for task '{task_name}'"
            logger.error(err)
            if job_id:
                self.report_status(job_id, status="failed", error=err)
            return

        func = self.registry[task_name]
        if job_id:
            self.report_status(job_id, status="running")

        try:
            result = func(*args, **kwargs)
            logger.info(f"Job {job_id} ({task_name}) completed successfully.")
            if job_id:
                self.report_status(job_id, status="completed", result=result)
        except Exception as exc:
            logger.exception(f"Job {job_id} ({task_name}) failed with exception: {exc}")
            if job_id:
                self.report_status(job_id, status="failed", error=str(exc))

    def start(self) -> None:
        """
        Begins polling the queue and executing registered tasks.
        """
        self._running = True
        logger.info(f"FlowForge Worker started for queue '{self.queue}' on {self.host}...")
        try:
            while self._running:
                try:
                    job = self.poll_job()
                    if job:
                        self.execute_job(job)
                    else:
                        time.sleep(self.poll_interval)
                except Exception as loop_err:
                    logger.error(f"Error in worker polling loop: {loop_err}")
                    time.sleep(self.poll_interval)
        except KeyboardInterrupt:
            logger.info("Worker received interrupt, stopping gracefully...")
        finally:
            self._running = False

    def stop(self) -> None:
        """
        Stops the worker polling loop.
        """
        self._running = False
