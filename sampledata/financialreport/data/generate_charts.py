#!/usr/bin/env python3
"""
Generate Bar Chart and Pie Chart for Annual Financial Report
"""

import matplotlib.pyplot as plt
import io
import base64
import os

def generate_bar_chart(filename: str) -> str:
    """Generate a bar chart and save as PNG, return base64 string"""
    # Dummy Data
    quarters = ['Q1', 'Q2', 'Q3', 'Q4']
    revenue = [150000, 180000, 160000, 210000]
    expenses = [100000, 110000, 105000, 130000]
    profit = [r - e for r, e in zip(revenue, expenses)]

    x = range(len(quarters))
    width = 0.25

    fig, ax = plt.subplots(figsize=(8, 5))
    rects1 = ax.bar([i - width for i in x], revenue, width, label='Revenue', color='#4CAF50')
    rects2 = ax.bar(x, expenses, width, label='Expenses', color='#F44336')
    rects3 = ax.bar([i + width for i in x], profit, width, label='Profit', color='#2196F3')

    ax.set_ylabel('Amount ($)')
    ax.set_title('Annual Financial Performance by Quarter')
    ax.set_xticks(x)
    ax.set_xticklabels(quarters)
    ax.legend()

    fig.tight_layout()
    
    # Save to file
    plt.savefig(filename)
    print(f"Bar Chart saved to: {filename}")
    
    # Convert to base64
    buffer = io.BytesIO()
    plt.savefig(buffer, format='PNG')
    base64_str = base64.b64encode(buffer.getvalue()).decode('utf-8')
    plt.close()
    
    return base64_str

def generate_pie_chart(filename: str) -> str:
    """Generate a pie chart and save as PNG, return base64 string"""
    # Dummy Data
    labels = ['R&D', 'Marketing', 'Operations', 'Salaries', 'Misc']
    sizes = [25, 20, 15, 35, 5]
    colors = ['#ff9999','#66b3ff','#99ff99','#ffcc99', '#c2c2f0']
    explode = (0.1, 0, 0, 0, 0)  # explode 1st slice

    fig1, ax1 = plt.subplots(figsize=(6, 6))
    ax1.pie(sizes, explode=explode, labels=labels, colors=colors, autopct='%1.1f%%',
            shadow=True, startangle=90)
    ax1.axis('equal')  # Equal aspect ratio ensures that pie is drawn as a circle.
    plt.title("Annual Expense Breakdown")

    # Save to file
    plt.savefig(filename)
    print(f"Pie Chart saved to: {filename}")
    
    # Convert to base64
    buffer = io.BytesIO()
    plt.savefig(buffer, format='PNG')
    base64_str = base64.b64encode(buffer.getvalue()).decode('utf-8')
    plt.close()
    
    return base64_str

def main():
    # Get the directory of this script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Generate Bar Chart
    bar_filename = os.path.join(script_dir, "bar_chart.png")
    bar_base64 = generate_bar_chart(bar_filename)
    
    # Generate Pie Chart
    pie_filename = os.path.join(script_dir, "pie_chart.png")
    pie_base64 = generate_pie_chart(pie_filename)
    
    print("\n" + "="*60)
    print("BASE64 ENCODED IMAGES GENERATED")
    print("="*60)
    
    # Save base64 strings to text files
    with open(os.path.join(script_dir, "bar_chart_base64.txt"), 'w') as f:
        f.write(bar_base64)
    print(f"\nBar Chart base64 saved to: bar_chart_base64.txt")
    
    with open(os.path.join(script_dir, "pie_chart_base64.txt"), 'w') as f:
        f.write(pie_base64)
    print(f"Pie Chart base64 saved to: pie_chart_base64.txt")

if __name__ == "__main__":
    main()
