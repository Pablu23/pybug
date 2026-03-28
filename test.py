import random
from typing import Any, Dict, List


class Employee:
    def __init__(self, name: str, age: int, skills: List[str]):
        self.name = name
        self.age = age
        self.skills = skills
        self.projects: Dict[str, Dict[str, Any]] = {}

    def add_project(self, project_name: str, details: Dict[str, Any]):
        self.projects[project_name] = details

    def __repr__(self):
        return f"<Employee {self.name}, age={self.age}, skills={self.skills}>"


class Company:
    def __init__(self, name: str):
        self.name = name
        self.employees: List[Employee] = []

    def hire_employee(self, employee: Employee):
        self.employees.append(employee)

    def list_employees(self):
        return [e.name for e in self.employees]

    def employee_summary(self) -> Dict[str, Any]:
        summary = {}
        for e in self.employees:
            summary[e.name] = {"age": e.age, "skills": e.skills, "projects": e.projects}
        return summary


def random_employee() -> Employee:
    names = ["Alice", "Bob", "Charlie", "Diana", "Eve", "Frank"]
    skills_pool = ["Python", "Go", "Rust", "JavaScript", "SQL", "Docker"]
    name = random.choice(names)
    age = random.randint(22, 45)
    skills = random.sample(skills_pool, k=random.randint(1, 4))
    return Employee(name, age, skills)


def generate_company(name: str, num_employees: int) -> Company:
    company = Company(name)
    for _ in range(num_employees):
        emp = random_employee()
        # assign some random projects
        for pid in range(random.randint(1, 3)):
            proj_name = f"Project-{pid+1}"
            emp.add_project(
                proj_name,
                {
                    "budget": random.randint(1000, 10000),
                    "deadline": f"2026-0{random.randint(1,9)}-{random.randint(10,28)}",
                    "tasks": [f"Task-{i+1}" for i in range(random.randint(1, 5))],
                },
            )
        company.hire_employee(emp)
    return company


def main():
    my_company = generate_company("TechCorp", 5)

    print("All employees:")
    for emp in my_company.employees:
        print(emp)

    summary = my_company.employee_summary()
    print("\nEmployee Summary Dictionary:")
    for name, info in summary.items():
        print(f"{name}: {info}")

    # Access nested structures
    print("\nAccessing nested info:")
    for emp in my_company.employees:
        for proj, details in emp.projects.items():
            print(
                f"{emp.name}'s {proj} budget is {details['budget']} with tasks {details['tasks']}"
            )

    # Filter employees with Python skill
    pythonistas = [e.name for e in my_company.employees if "Python" in e.skills]
    print(f"\nEmployees with Python skill: {pythonistas}")

    # Update project budget for first employee
    if my_company.employees:
        first = my_company.employees[0]
        for proj in first.projects:
            first.projects[proj]["budget"] += 500
        print(f"\nUpdated {first.name}'s projects: {first.projects}")


if __name__ == "__main__":
    main()
