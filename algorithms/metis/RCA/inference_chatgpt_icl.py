
import time
import requests
import pandas as pd
from read_data import get_one_prompt
from read_data import get_instrucion
import concurrent.futures

#your openai key
Authorization = ""

# prepare ICL examples:
def get_example(task_name, lineID=None):
    example1 = ""
    example_label1 = ""
    example2 = ""
    example_label2 = ""
    example3 = ""
    example_label3 = ""
    example_prompt = ""

    if task_name == "task5":
        example1 = "public class A { <line0> @AfterClass <line1> public static void afterClass() { <line2> try { <line3> standaloneConsul.stop(); <line4> } catch (Exception e) { <line5> } <line6> } <line7> } <line8>"
        example_label1 = "<line5>      logger.info(""Failed to stop standalone consul"");"
        example2 = "public class A { <line0> @Override <line1> public void onError(FacebookException error) { <line2> } <line3> } <line4>"
        example_label2 = "<line2>    logger.error(error);"
        example3 = "public class A { <line0> public void finish() throws IOException { <line1> this.data.rewind(); <line2> try { <line3> pieceStorage.savePiece(index, this.data.array()); <line4> } finally { <line5> this.data = null; <line6> } <line7> } <line8> } <line9>"
        example_label3 = "<line2>    logger.trace(""Recording {}..."", this);"
    example_prompt = "Example: " + example1 + "\n Label: " + example_label1 + "\n \n" + "Example: " + example2 + "\n Label: " + example_label2 + "\n \n" + "Example: " + example3 + "\n Label: " + example_label3 + "\n \n"
    return example_prompt



def chat_with_gpt(instruction,input):
    url = "https://api.openai.com/v1/chat/completions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": Authorization
    }

    data = {
        "model": "gpt-4",
        # change model here
        #"model": "gpt-3.5-turbo", 
        "messages": [
            {
                "role": "system",
                "content": instruction
            },
            {
                "role": "user",
                "content": input
            }
        ]
    }
    # print("data: \n",data)
    response = requests.post(url, json=data, headers=headers)
    output = response.json()

    return output


output_list = []
query_list = []
task_list = []
groundtruth_list = []

data_path = "./task1-5/mixtasks1-4_test.tsv"
result_path = "./task1-5/chatgpt4_result.tsv"

df = pd.read_csv(data_path, sep='\t')
list_data_dict = df.to_dict('records')

count = 0
batch_count = 0
batch_size = 5  # Set batch size
max_retries = 3  # Set maximum number of retries

for row in list_data_dict:
    instruction = get_instrucion(row['task'])
    example_prompt = get_example(row['task'])
    input_text = example_prompt + "Query: " + row['code'] + "\n" + "Label:"
    retries = 0
    response = None

    while retries < max_retries:
        response = chat_with_gpt(instruction, input_text)
        if response and 'choices' in response and response['choices'] and 'message' in response['choices'][0] and 'content' in response['choices'][0]['message']:
            break
        retries += 1
        time.sleep(3)
        print(f"Retrying request {retries} for row {count}")
    
    if not response or 'choices' not in response or not response['choices'] or 'message' not in response['choices'][0] or 'content' not in response['choices'][0]['message']:
        print(f"Failed to get valid response for row {count}")
        continue
    
    groundtruth = row['label']
    count += 1
    print(count)
    output = repr(response['choices'][0]['message']['content'])

    output_list.append(output)
    query_list.append(row['code'])
    task_list.append(row['task'])
    groundtruth_list.append(groundtruth)

    batch_count += 1
    if batch_count == batch_size:
        df_result = pd.DataFrame({'task': task_list, 'prompt': query_list, 'label': groundtruth_list, 'predict': output_list})
        df_result.to_csv(result_path, sep='\t', index=False, mode='a', header=False)
        
        # Delete processed batch from original data
        df = df.drop(df.index[:batch_count])
        df.to_csv(data_path, sep='\t', index=False)

        # Reset lists and batch count
        output_list = []
        query_list = []
        task_list = []
        groundtruth_list = []
        batch_count = 0

if batch_count > 0:
    df_result = pd.DataFrame({'task': task_list, 'prompt': query_list, 'label': groundtruth_list, 'predict': output_list})
    df_result.to_csv(result_path, sep='\t', index=False, mode='a', header=False)
    df = df.drop(df.index[:batch_count])
    df.to_csv(data_path, sep='\t', index=False)

print("done")

