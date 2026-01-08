"""
File-based locking utility for preventing concurrent access to shared resources.

Provides file locks for:
- Task assignment operations (prevent two agents from taking same task)
- State file updates (prevent concurrent writes)
- Atomic read-modify-write operations

Uses OS-level file locking (fcntl on Unix, msvcrt on Windows) for cross-process coordination.
"""

import fcntl
import logging
import os
import platform
import time
from contextlib import contextmanager
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)

# Windows support
if platform.system() == "Windows":
    try:
        import msvcrt
        HAS_WINDOWS_LOCK = True
    except ImportError:
        HAS_WINDOWS_LOCK = False
else:
    HAS_WINDOWS_LOCK = False


class FileLock:
    """
    File-based lock for coordinating access to shared files.
    
    Uses OS-level file locking to prevent concurrent access across processes.
    Thread-safe and process-safe.
    """

    def __init__(self, lock_file: Path, timeout: float = 10.0, poll_interval: float = 0.1):
        """
        Initialize file lock.

        Args:
            lock_file: Path to lock file (will be created if needed)
            timeout: Maximum time to wait for lock (seconds)
            poll_interval: How often to retry acquiring lock (seconds)
        """
        self.lock_file = Path(lock_file)
        self.timeout = timeout
        self.poll_interval = poll_interval
        self.lock_fd: Optional[int] = None
        self.is_windows = platform.system() == "Windows"

    def acquire(self, blocking: bool = True) -> bool:
        """
        Acquire the lock.

        Args:
            blocking: If True, wait for lock; if False, return immediately if locked

        Returns:
            True if lock acquired, False otherwise
        """
        # Ensure lock file directory exists
        self.lock_file.parent.mkdir(parents=True, exist_ok=True)

        try:
            # Open lock file (create if doesn't exist)
            self.lock_fd = os.open(
                str(self.lock_file),
                os.O_CREAT | os.O_RDWR
            )

            start_time = time.time()

            while True:
                try:
                    if self.is_windows and HAS_WINDOWS_LOCK:
                        # Windows: use msvcrt locking
                        msvcrt.locking(self.lock_fd, msvcrt.LK_NBLCK, 1)
                    else:
                        # Unix: use fcntl
                        fcntl.flock(self.lock_fd, fcntl.LOCK_EX | fcntl.LOCK_NB)

                    # Lock acquired
                    logger.debug(f"Lock acquired: {self.lock_file}")
                    return True

                except (IOError, OSError) as e:
                    # Lock is held by another process
                    if not blocking:
                        os.close(self.lock_fd)
                        self.lock_fd = None
                        return False

                    # Check timeout
                    if time.time() - start_time > self.timeout:
                        logger.warning(f"Lock timeout after {self.timeout}s: {self.lock_file}")
                        os.close(self.lock_fd)
                        self.lock_fd = None
                        return False

                    # Wait and retry
                    time.sleep(self.poll_interval)

        except Exception as e:
            logger.error(f"Error acquiring lock {self.lock_file}: {e}")
            if self.lock_fd is not None:
                try:
                    os.close(self.lock_fd)
                except:
                    pass
                self.lock_fd = None
            return False

    def release(self) -> None:
        """Release the lock."""
        if self.lock_fd is not None:
            try:
                if self.is_windows and HAS_WINDOWS_LOCK:
                    msvcrt.locking(self.lock_fd, msvcrt.LK_UNLCK, 1)
                else:
                    fcntl.flock(self.lock_fd, fcntl.LOCK_UN)

                os.close(self.lock_fd)
                self.lock_fd = None
                logger.debug(f"Lock released: {self.lock_file}")
            except Exception as e:
                logger.error(f"Error releasing lock {self.lock_file}: {e}")

    def __enter__(self):
        """Context manager entry."""
        if not self.acquire():
            raise RuntimeError(f"Failed to acquire lock: {self.lock_file}")
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.release()
        return False

    def __del__(self):
        """Cleanup on deletion."""
        self.release()


@contextmanager
def task_lock(task_id: Optional[str] = None, timeout: float = 10.0):
    """
    Context manager for locking task operations.

    Locks the entire state file (if task_id is None) or a specific task.
    Prevents concurrent modifications by multiple agents.

    Args:
        task_id: Optional task ID to lock (None = lock entire state file)
        timeout: Maximum time to wait for lock (seconds)

    Example:
        with task_lock(task_id="T-123"):
            # Atomically check and assign task
            state = _load_todo2_state()
            task = find_task(state, "T-123")
            if not task.get("assignee"):
                task["assignee"] = {"name": "agent-1"}
                _save_todo2_state(state)
    """
    from .project_root import find_project_root

    project_root = find_project_root()

    if task_id:
        # Task-specific lock file
        lock_file = project_root / ".todo2" / "locks" / f"task_{task_id}.lock"
    else:
        # Global state file lock
        lock_file = project_root / ".todo2" / "state.todo2.json.lock"

    lock = FileLock(lock_file, timeout=timeout)
    try:
        if not lock.acquire(blocking=True):
            raise RuntimeError(f"Failed to acquire lock for task {task_id or 'state file'}")
        yield lock
    finally:
        lock.release()


@contextmanager
def state_file_lock(timeout: float = 10.0):
    """
    Context manager for locking the entire Todo2 state file.

    Use this for operations that modify multiple tasks or the state structure.

    Args:
        timeout: Maximum time to wait for lock (seconds)

    Example:
        with state_file_lock():
            state = _load_todo2_state()
            # Modify state...
            _save_todo2_state(state)
    """
    with task_lock(task_id=None, timeout=timeout):
        yield


def try_lock_task(task_id: str, timeout: float = 0.0) -> Optional[FileLock]:
    """
    Try to acquire a lock for a specific task (non-blocking).

    Args:
        task_id: Task ID to lock
        timeout: Timeout (0 = immediate return)

    Returns:
        FileLock instance if acquired, None otherwise

    Example:
        lock = try_lock_task("T-123")
        if lock:
            try:
                # Work on task...
            finally:
                lock.release()
    """
    from .project_root import find_project_root

    project_root = find_project_root()
    lock_file = project_root / ".todo2" / "locks" / f"task_{task_id}.lock"
    lock = FileLock(lock_file, timeout=timeout)

    if lock.acquire(blocking=(timeout > 0)):
        return lock
    return None
