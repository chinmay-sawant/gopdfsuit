import random
import matplotlib.pyplot as plt
import os

def generate_complex_calculus():
    """Generates a long and complex calculus formula involving integration and differentiation."""
    
    # Components for differentiation
    diff_terms = [
        r"\frac{d}{dx} \left( x^3 \sin(ax^2) + e^{-bx} \right)",
        r"\frac{\partial f}{\partial x} \sqrt{x^2 + y^2}",
        r"\frac{d^2y}{dx^2} + p(x)\frac{dy}{dx} + q(x)y",
        r"\nabla \cdot \mathbf{E} = \frac{\rho}{\epsilon_0}",
        r"\frac{d}{dt} \oint \mathbf{B} \cdot d\mathbf{A}"
    ]
    
    # Components for integration
    int_terms = [
        r"\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}",
        r"\oint_{C} P dx + Q dy = \iint_{D} \left( \frac{\partial Q}{\partial x} - \frac{\partial P}{\partial y} \right) dA",
        r"\int \frac{1}{\sqrt{a^2 - x^2}} dx = \arcsin\left(\frac{x}{a}\right) + C",
        r"\lim_{n \to \infty} \sum_{i=1}^{n} f(x_i^*) \Delta x = \int_{a}^{b} f(x) dx"
    ]
    
    # Random constants
    a = random.randint(2, 9)
    b = random.randint(1, 5)
    
    # Construct a "Long" formula by combining parts
    part1 = random.choice(diff_terms).replace('a', str(a)).replace('b', str(b))
    part2 = random.choice(int_terms)
    
    # Add some additional complexity
    formula = fr"$ {part1} \implies \int \left( {part2} \right) dx = \sum_{{k=1}}^{{\infty}} \frac{{1}}{{k^{{a}}}} $"
    
    return formula

def generate_definite_integral_problem():
    """Generates a long definite integral problem."""
    a = random.randint(0, 2)
    b = random.randint(3, 10)
    c = random.randint(1, 5)
    
    formula = (
        r"$\int_{" + str(a) + r"}^{" + str(b) + r"} \left( \frac{x^2 + " + str(c) + r"x}{\sqrt{x^2 + 1}} + e^{" + str(a) + r"x} \cos(" + str(c) + r"x) \right) dx = ?$"
    )
    return formula

def generate_derivative_chain_problem():
    """Generates a long derivative problem using the chain rule."""
    a = random.randint(2, 5)
    b = random.randint(2, 5)
    
    formula = (
        r"$\frac{d}{dx} \left[ \ln\left( \sin\left( e^{" + str(a) + r"x^2 + " + str(b) + r"x} \right) \right) + \frac{x^3}{\tan(x)} \right] = ?$"
    )
    return formula

def save_math_svg(math_text, filename="problem.svg"):
    """Renders LaTeX math text to an SVG file."""
    # Use Matplotlib's mathtext engine for standard math rendering
    plt.rcParams.update({
        "text.usetex": False,
        "svg.fonttype": "path", # Important for portability
        "figure.facecolor": "none",
        "axes.facecolor": "none",
    })

    # Create figure with dynamic size based on text length (rough estimation)
    # The font size is 32, so we adjust the figure size to avoid clipping
    fig = plt.figure(figsize=(12, 3))
    
    # Add text to the figure. 
    fig.text(0.5, 0.5, math_text, fontsize=28, ha='center', va='center')

    # Save as SVG. 
    plt.savefig(filename, format='svg', bbox_inches='tight', pad_inches=0.2, transparent=True)
    plt.close(fig)
    print(f"Successfully generated complex math SVG: {filename}")
    print(f"Mathematical expression: {math_text}")

if __name__ == "__main__":
    # Choose a random problem type from the new calculus generators
    problem_generators = [
        generate_complex_calculus,
        generate_definite_integral_problem,
        generate_derivative_chain_problem
    ]
    
    selected_problem = random.choice(problem_generators)()
    output_path = "math_problem.svg"
    
    save_math_svg(selected_problem, output_path)
    
    print(f"Saved to: {os.path.abspath(output_path)}")
