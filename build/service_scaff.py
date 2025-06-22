import os
import sys

# === Configuration ===
TEMPLATE_FILE_PATH = "../api/v1/service_template.txt"  # Replace with your template file path
OUTPUT_FILENAME = "service.go"               # Name of the output file in the new directory
PLACEHOLDER = "Template"                  # Placeholder text to replace in template


def scaffold_directory(name: str, service: str):
    new_dir = os.path.join("../api", name)
    # Check if directory already exists
    if os.path.exists(new_dir):
        print(f"Error: Directory '{new_dir}' already exists.")
        sys.exit(1)

    # Create the new directory
    os.makedirs(new_dir)
    print(f"Created directory: {new_dir}")

    # Read the template file
    try:
        with open(TEMPLATE_FILE_PATH, "r", encoding="utf-8") as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: Template file '{TEMPLATE_FILE_PATH}' not found.")
        sys.exit(1)

    # Replace the placeholder with the service name
    content = content.replace(PLACEHOLDER, service)

    # Write the new file inside the new directory
    output_path = os.path.join(new_dir, OUTPUT_FILENAME)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(content)

    print(f"Created file: {output_path}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("""
        Usage: python scaffold.py <service-name> <struct-prefix>

        e.g
        service-name of 'admin' with struct-prefix of `Admin`
        results in a directory within 'api/' -> 'admin/service.go'
        'service.go' contains 'AdminService'
        this service implements 'api/v1/service.go'
        """)
        sys.exit(1)

    dir_name = sys.argv[1]
    service_name = sys.argv[2]

    scaffold_directory(dir_name, service_name)
