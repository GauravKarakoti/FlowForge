class Client:
    def __init__(self, host="http://localhost:8080", token=None):
        self.host = host
        self.token = token
        # TODO: Initialize HTTP session

    def enqueue(self, task_name, args=None, kwargs=None, retry_limit=3):
        """
        Submits a job to the FlowForge control plane.
        """
        # TODO: Serialize payload and make POST request to control plane
        pass


class Worker:
    def __init__(self, queue="default"):
        self.queue = queue
        self.registry = {}
        # TODO: Setup connection to broker or control plane

    def register_task(self, name, func):
        """
        Registers a local python function to handle a specific task name.
        """
        # TODO: Add function to self.registry mapping
        pass

    def start(self):
        """
        Begins polling the queue and executing registered tasks.
        """
        # TODO: Implement infinite polling loop
        # TODO: Handle execution, exceptions, and state reporting
        pass