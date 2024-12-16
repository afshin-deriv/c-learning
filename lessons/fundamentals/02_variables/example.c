#include <stdio.h>

int main() {
    // Declaring and initializing variables
    int age = 25;
    float height = 1.75;
    
    // Performing calculations
    float bmi_weight = 70.5;  // weight in kg
    float bmi = bmi_weight / (height * height);
    
    // Printing values with appropriate formats
    printf("Age: %d years\n", age);
    printf("Height: %.2f meters\n", height);
    printf("Weight: %.1f kg\n", bmi_weight);
    printf("BMI: %.1f\n", bmi);
    
    return 0;
}