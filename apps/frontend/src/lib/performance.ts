// Performance monitoring utilities
export const measurePageLoad = (pageName: string) => {
  if (typeof window === 'undefined') return;
  
  // Use Performance API to measure page load time
  const perfData = window.performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
  
  if (perfData) {
    const loadTime = perfData.loadEventEnd - perfData.fetchStart;
    console.log(`[Performance] ${pageName} total load time: ${loadTime.toFixed(2)}ms`);
    
    // Log detailed metrics in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`[Performance] ${pageName} metrics:`, {
        dns: (perfData.domainLookupEnd - perfData.domainLookupStart).toFixed(2) + 'ms',
        tcp: (perfData.connectEnd - perfData.connectStart).toFixed(2) + 'ms',
        request: (perfData.responseStart - perfData.requestStart).toFixed(2) + 'ms',
        response: (perfData.responseEnd - perfData.responseStart).toFixed(2) + 'ms',
        domParsing: (perfData.domInteractive - perfData.domLoading).toFixed(2) + 'ms',
        domContentLoaded: (perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart).toFixed(2) + 'ms',
        onLoad: (perfData.loadEventEnd - perfData.loadEventStart).toFixed(2) + 'ms',
      });
    }
  }
};

// Measure component render time
export const measureRender = (componentName: string, callback: () => void) => {
  if (typeof window === 'undefined') return callback();
  
  const startTime = performance.now();
  callback();
  const endTime = performance.now();
  
  if (process.env.NODE_ENV === 'development') {
    console.log(`[Performance] ${componentName} render time: ${(endTime - startTime).toFixed(2)}ms`);
  }
};

// Debounce utility for expensive operations
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) => {
  let timeoutId: NodeJS.Timeout;
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func(...args), delay);
  };
};

// Throttle utility for frequent operations
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

// Lazy load images with Intersection Observer
export const lazyLoadImages = () => {
  if (typeof window === 'undefined' || !('IntersectionObserver' in window)) return;
  
  const images = document.querySelectorAll('img[data-src]');
  
  const imageObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement;
        img.src = img.dataset.src || '';
        img.removeAttribute('data-src');
        imageObserver.unobserve(img);
      }
    });
  });
  
  images.forEach(img => imageObserver.observe(img));
};

// Prefetch critical resources
export const prefetchResources = (urls: string[]) => {
  if (typeof window === 'undefined') return;
  
  urls.forEach(url => {
    const link = document.createElement('link');
    link.rel = 'prefetch';
    link.href = url;
    document.head.appendChild(link);
  });
};

// Cache API responses
const apiCache = new Map<string, { data: any; timestamp: number }>();
const CACHE_TTL = 5 * 60 * 1000; // 5 minutes

export const getCachedData = <T>(key: string): T | null => {
  const cached = apiCache.get(key);
  
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.data as T;
  }
  
  apiCache.delete(key);
  return null;
};

export const setCachedData = <T>(key: string, data: T): void => {
  apiCache.set(key, { data, timestamp: Date.now() });
};

// Clear old cache entries
export const clearExpiredCache = () => {
  const now = Date.now();
  
  apiCache.forEach((value, key) => {
    if (now - value.timestamp >= CACHE_TTL) {
      apiCache.delete(key);
    }
  });
};

// Schedule cache cleanup
if (typeof window !== 'undefined') {
  setInterval(clearExpiredCache, 60 * 1000); // Run every minute
}