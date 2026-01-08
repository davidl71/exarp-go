"""
Unified JSON Caching Utility

Provides centralized, reusable JSON file caching with:
- File modification time (mtime) invalidation
- TTL (time-to-live) expiration
- LRU eviction for cache size limits
- Decorator and context manager patterns
- Cache statistics and monitoring
- Thread-safe operations

Usage:
    # Decorator pattern
    @json_file_cache(file_path=Path(".todo2/commits.json"), ttl=300)
    def load_commits() -> list[dict]:
        with open(".todo2/commits.json") as f:
            return json.load(f).get("commits", [])

    # Context manager pattern
    cache = JsonFileCache(Path(".todo2/commits.json"), ttl=300)
    data = cache.get_or_load()
"""

import json
import logging
import threading
import time
from collections import OrderedDict
from functools import wraps
from pathlib import Path
from typing import Any, Callable, Optional, TypeVar, Union

logger = logging.getLogger(__name__)

T = TypeVar('T')

# Maximum number of file caches to maintain (LRU eviction when exceeded)
MAX_CACHES = 100


class JsonFileCache:
    """
    File-based JSON cache with mtime invalidation and TTL support.
    
    Thread-safe caching for JSON files with automatic invalidation
    based on file modification time and optional TTL expiration.
    """

    def __init__(
        self,
        file_path: Union[Path, str],
        ttl: Optional[int] = None,
        max_size: Optional[int] = None,
        enable_stats: bool = True,
        default_value: Any = None,
        raise_on_error: bool = False,
    ):
        """
        Initialize JSON file cache.

        Args:
            file_path: Path to JSON file to cache
            ttl: Time-to-live in seconds (None = mtime only, no TTL)
            max_size: Maximum cache entries (None = unlimited, not used for single file)
            enable_stats: Enable statistics collection
            default_value: Value to return on error (if raise_on_error=False)
            raise_on_error: Whether to raise exceptions or return default_value
        """
        self.file_path = Path(file_path) if isinstance(file_path, str) else file_path
        self.ttl = ttl
        self.max_size = max_size  # Reserved for future multi-file caching
        self.enable_stats = enable_stats
        self.default_value = default_value if default_value is not None else {}
        self.raise_on_error = raise_on_error

        # Cache state
        self._cache: Optional[dict[str, Any]] = None
        self._cache_mtime: Optional[float] = None
        self._cache_timestamp: Optional[float] = None  # For TTL

        # Thread safety
        self._lock = threading.Lock()

        # Statistics
        self._stats = {
            "hits": 0,
            "misses": 0,
            "invalidations": 0,
            "invalidations_mtime": 0,
            "invalidations_ttl": 0,
            "errors": 0,
        } if enable_stats else None

    def _is_valid(self) -> bool:
        """Check if cache is valid (not expired, file not modified)."""
        if self._cache is None:
            return False

        # Check TTL expiration
        if self.ttl is not None and self._cache_timestamp is not None:
            if time.time() - self._cache_timestamp > self.ttl:
                if self.enable_stats:
                    self._stats["invalidations_ttl"] += 1
                return False

        # Check file modification time
        try:
            if not self.file_path.exists():
                return False

            current_mtime = self.file_path.stat().st_mtime
            if self._cache_mtime is None or self._cache_mtime != current_mtime:
                if self.enable_stats and self._cache_mtime is not None:
                    self._stats["invalidations_mtime"] += 1
                return False

            return True
        except (OSError, IOError) as e:
            logger.warning(f"Error checking file mtime for {self.file_path}: {e}")
            return False

    def get(self) -> Optional[dict[str, Any]]:
        """
        Get cached data if valid, None if expired/missing.

        Returns:
            Cached data dict if valid, None otherwise
        """
        with self._lock:
            if self._is_valid():
                if self.enable_stats:
                    self._stats["hits"] += 1
                return self._cache
            else:
                if self.enable_stats:
                    self._stats["misses"] += 1
                return None

    def get_or_load(self) -> dict[str, Any]:
        """
        Get cached data or load from file.

        Returns:
            Data dict from cache or file
        """
        with self._lock:
            # Try cache first
            cached = self.get()
            if cached is not None:
                return cached

            # Load from file
            try:
                if not self.file_path.exists():
                    if self.raise_on_error:
                        raise FileNotFoundError(f"File not found: {self.file_path}")
                    logger.warning(f"File not found: {self.file_path}, returning default")
                    return self.default_value

                # Load JSON
                with open(self.file_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)

                # Update cache
                self._cache = data
                self._cache_timestamp = time.time()
                try:
                    self._cache_mtime = self.file_path.stat().st_mtime
                except (OSError, IOError) as e:
                    logger.warning(f"Error getting mtime for {self.file_path}: {e}")
                    self._cache_mtime = None

                if self.enable_stats:
                    self._stats["misses"] += 1

                return data

            except json.JSONDecodeError as e:
                if self.enable_stats:
                    self._stats["errors"] += 1
                logger.error(f"Invalid JSON in {self.file_path}: {e}")
                if self.raise_on_error:
                    raise
                return self.default_value

            except (OSError, IOError) as e:
                if self.enable_stats:
                    self._stats["errors"] += 1
                logger.error(f"Error reading {self.file_path}: {e}")
                if self.raise_on_error:
                    raise
                return self.default_value

    def invalidate(self) -> None:
        """Manually invalidate cache."""
        with self._lock:
            self._cache = None
            self._cache_mtime = None
            self._cache_timestamp = None
            if self.enable_stats:
                self._stats["invalidations"] += 1

    def clear(self) -> None:
        """Clear all cache data and reset statistics."""
        with self._lock:
            self.invalidate()
            if self.enable_stats:
                self._stats = {
                    "hits": 0,
                    "misses": 0,
                    "invalidations": 0,
                    "invalidations_mtime": 0,
                    "invalidations_ttl": 0,
                    "errors": 0,
                }

    def get_stats(self) -> dict[str, Any]:
        """
        Get cache statistics.

        Returns:
            Dictionary with cache statistics
        """
        if not self.enable_stats:
            return {"enabled": False}

        with self._lock:
            total_requests = self._stats["hits"] + self._stats["misses"]
            hit_rate = (
                self._stats["hits"] / total_requests
                if total_requests > 0
                else 0.0
            )

            return {
                "hits": self._stats["hits"],
                "misses": self._stats["misses"],
                "hit_rate": round(hit_rate, 3),
                "invalidations": self._stats["invalidations"],
                "invalidations_mtime": self._stats["invalidations_mtime"],
                "invalidations_ttl": self._stats["invalidations_ttl"],
                "errors": self._stats["errors"],
                "ttl": self.ttl,
                "file_path": str(self.file_path),
                "cached": self._cache is not None,
            }


class JsonCacheManager:
    """
    Global cache manager for multiple JSON file caches.

    Singleton pattern for managing multiple file caches across the application.
    Uses LRU (Least Recently Used) eviction when cache limit is reached.
    """

    _instance: Optional['JsonCacheManager'] = None
    _lock = threading.Lock()

    def __init__(self, max_caches: int = MAX_CACHES):
        """
        Initialize cache manager.
        
        Args:
            max_caches: Maximum number of file caches to maintain (default: MAX_CACHES)
        """
        self._caches: OrderedDict[str, JsonFileCache] = OrderedDict()
        self._manager_lock = threading.Lock()
        self._max_caches = max_caches

    @classmethod
    def get_instance(cls) -> 'JsonCacheManager':
        """Get singleton instance of cache manager."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = cls()
        return cls._instance

    def get_cache(
        self,
        file_path: Union[Path, str],
        ttl: Optional[int] = None,
        max_size: Optional[int] = None,
        enable_stats: bool = True,
    ) -> JsonFileCache:
        """
        Get or create cache for file with LRU eviction.

        When cache limit is reached, the least recently used cache is evicted.
        Accessed caches are moved to the end (most recently used).

        Args:
            file_path: Path to JSON file
            ttl: Time-to-live in seconds
            max_size: Maximum cache size (reserved for future)
            enable_stats: Enable statistics

        Returns:
            JsonFileCache instance for the file
        """
        path_str = str(Path(file_path).absolute())

        with self._manager_lock:
            if path_str in self._caches:
                # Cache exists - move to end (most recently used) for LRU
                cache = self._caches.pop(path_str)
                self._caches[path_str] = cache
                return cache
            
            # Cache doesn't exist - create new one
            # Evict oldest cache if at limit (LRU eviction)
            if len(self._caches) >= self._max_caches:
                oldest_path, oldest_cache = self._caches.popitem(last=False)
                logger.debug(
                    f"Evicting LRU cache (limit {self._max_caches} reached): {oldest_path}"
                )
                # Invalidate the evicted cache to free memory
                oldest_cache.invalidate()
            
            # Create and add new cache (will be at end = most recently used)
            new_cache = JsonFileCache(
                file_path=file_path,
                ttl=ttl,
                max_size=max_size,
                enable_stats=enable_stats,
            )
            self._caches[path_str] = new_cache
            return new_cache

    def invalidate_file(self, file_path: Union[Path, str]) -> None:
        """Invalidate cache for specific file."""
        path_str = str(Path(file_path).absolute())
        with self._manager_lock:
            if path_str in self._caches:
                self._caches[path_str].invalidate()

    def invalidate_all(self) -> None:
        """Invalidate all caches."""
        with self._manager_lock:
            for cache in self._caches.values():
                cache.invalidate()

    def get_all_stats(self) -> dict[str, dict[str, Any]]:
        """
        Get statistics for all caches.

        Returns:
            Dictionary mapping file paths to their statistics
        """
        with self._manager_lock:
            stats = {
                path: cache.get_stats()
                for path, cache in self._caches.items()
            }
            # Add manager-level stats
            stats["_manager"] = {
                "total_caches": len(self._caches),
                "max_caches": self._max_caches,
                "cache_limit_reached": len(self._caches) >= self._max_caches,
            }
            return stats
    
    def clear_all(self) -> None:
        """
        Clear all caches and reset the cache manager.
        
        Useful for testing or memory cleanup.
        """
        with self._manager_lock:
            for cache in self._caches.values():
                cache.clear()
            self._caches.clear()


def json_file_cache(
    file_path: Union[Path, str],
    ttl: Optional[int] = None,
    max_size: Optional[int] = None,
    enable_stats: bool = True,
    default_value: Any = None,
    raise_on_error: bool = False,
) -> Callable[[Callable[[], T]], Callable[[], T]]:
    """
    Decorator for caching JSON file loads.

    Args:
        file_path: Path to JSON file
        ttl: Time-to-live in seconds (None = mtime only)
        max_size: Maximum cache size (reserved for future)
        enable_stats: Enable statistics
        default_value: Value to return on error
        raise_on_error: Whether to raise exceptions

    Returns:
        Decorated function with caching

    Example:
        @json_file_cache(file_path=Path(".todo2/commits.json"), ttl=300)
        def load_commits() -> list[dict]:
            with open(".todo2/commits.json") as f:
                return json.load(f).get("commits", [])
    """
    cache = JsonFileCache(
        file_path=file_path,
        ttl=ttl,
        max_size=max_size,
        enable_stats=enable_stats,
        default_value=default_value,
        raise_on_error=raise_on_error,
    )

    def decorator(func: Callable[[], T]) -> Callable[[], T]:
        @wraps(func)
        def wrapper() -> T:
            # Try cache first
            cached = cache.get()
            if cached is not None:
                return cached  # type: ignore

            # Call function and cache result
            result = func()
            with cache._lock:
                cache._cache = result  # type: ignore
                cache._cache_timestamp = time.time()
                try:
                    cache._cache_mtime = cache.file_path.stat().st_mtime
                except (OSError, IOError):
                    cache._cache_mtime = None
            return result

        # Attach cache to function for manual invalidation
        wrapper._cache = cache  # type: ignore
        return wrapper

    return decorator


def json_file_cache_context(
    file_path: Union[Path, str],
    ttl: Optional[int] = None,
    enable_stats: bool = True,
) -> JsonFileCache:
    """
    Context manager for JSON file caching.

    Args:
        file_path: Path to JSON file
        ttl: Time-to-live in seconds
        enable_stats: Enable statistics

    Returns:
        JsonFileCache instance (usable as context manager)

    Example:
        with json_file_cache_context(Path(".todo2/commits.json")) as cache:
            data = cache.get_or_load()
    """
    return JsonFileCache(
        file_path=file_path,
        ttl=ttl,
        enable_stats=enable_stats,
    )
