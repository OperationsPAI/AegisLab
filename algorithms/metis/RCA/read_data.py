def get_prompt(instruction, input):
    str1 = "Instruction: " + instruction + "\n"
    str2 = "Input: " + input + "\n"
    str3 = "Output:" 

    prompt = str1 + str2 + str3
    return prompt

def get_instrucion(task_name, lineID=None):
    instruction = ""
    # generate logging statement (pos,level,msg)
    if task_name == "task5":
        instruction = "Generate a complete log statement with an appropriate line index ahead for the given input code."
    return instruction


def read_data(row):
    task_name = row['task']
    input = row['code']
    output = row['label']
    lineID = row['lineID']

    prompt = []
    for i in range(len(input)):
        prompt.append(get_prompt(get_instrucion(task_name[i], lineID[i]), input[i]))
    
    return prompt, output


def get_one_prompt(row):
    task_name = row['task']
    input = row['code']
    lineID = row['lineID']

    prompt = get_prompt(get_instrucion(task_name, lineID), input)
    return prompt
