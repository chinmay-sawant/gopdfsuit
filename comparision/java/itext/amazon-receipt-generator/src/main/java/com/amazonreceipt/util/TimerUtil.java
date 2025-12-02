package com.amazonreceipt.util;

/**
 * Utility class for timing operations.
 */
public class TimerUtil {
    
    private long startTime;
    private long endTime;
    private String operationName;
    
    public TimerUtil(String operationName) {
        this.operationName = operationName;
    }
    
    /**
     * Start the timer.
     */
    public void start() {
        startTime = System.nanoTime();
        System.out.println("\n╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║  ⏱️  TIMER STARTED: " + padRight(operationName, 40) + "║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");
    }
    
    /**
     * Stop the timer and print results.
     */
    public void stop() {
        endTime = System.nanoTime();
        long durationNanos = endTime - startTime;
        double durationMillis = durationNanos / 1_000_000.0;
        double durationSeconds = durationNanos / 1_000_000_000.0;
        
        System.out.println("\n╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║  ⏱️  TIMER STOPPED: " + padRight(operationName, 40) + "║");
        System.out.println("╠══════════════════════════════════════════════════════════════╣");
        System.out.println("║  📊 Duration Results:                                        ║");
        System.out.println("║     • Nanoseconds:  " + padRight(String.format("%,d ns", durationNanos), 40) + "║");
        System.out.println("║     • Milliseconds: " + padRight(String.format("%.3f ms", durationMillis), 40) + "║");
        System.out.println("║     • Seconds:      " + padRight(String.format("%.6f s", durationSeconds), 40) + "║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");
    }
    
    /**
     * Get duration in milliseconds.
     */
    public double getDurationMillis() {
        return (endTime - startTime) / 1_000_000.0;
    }
    
    /**
     * Get duration in seconds.
     */
    public double getDurationSeconds() {
        return (endTime - startTime) / 1_000_000_000.0;
    }
    
    /**
     * Pad string to right with spaces.
     */
    private static String padRight(String s, int n) {
        if (s.length() >= n) {
            return s.substring(0, n);
        }
        return String.format("%-" + n + "s", s);
    }
}
