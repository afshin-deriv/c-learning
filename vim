{
    "lessons": [
        {
            "id": 1,
            "title": "Introduction to C Programming",
            "description": "Learn the basics of C programming, including how to write a simple program and understand its structure.",
            "learning_objectives": [
                "Understand the basic structure of a C program",
                "Learn about the main() function",
                "Write your first Hello World program",
                "Compile and run a C program"
            ],
            "example_code": "#include <stdio.h>\n\nint main() {\n    printf(\"Hello, World!\\n\");\n    return 0;\n}\n",
            "test_cases": [
                {
                    "input": "",
                    "expected_output": "Hello, World!\n",
                    "description": "Program should print 'Hello, World!' followed by a newline"
                }
            ],
            "prerequisites": []
        },
        {
            "id": 2,
            "title": "Variables and Data Types",
            "description": "Learn about basic data types in C and how to use variables.",
            "learning_objectives": [
                "Understand different data types (int, float, char)",
                "Learn how to declare and initialize variables",
                "Practice using variables in expressions",
                "Learn about type conversion"
            ],
            "example_code": "#include <stdio.h>\n\nint main() {\n    int age = 25;\n    float height = 1.75;\n    printf(\"Age: %d, Height: %.2f\\n\", age, height);\n    return 0;\n}\n",
            "test_cases": [
                {
                    "input": "",
                    "expected_output": "Age: 25, Height: 1.75\n",
                    "description": "Program should print age and height with correct formatting"
                }
            ],
            "prerequisites": [
                1
            ]
        }
    ]
}