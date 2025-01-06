from langchain.agents import Tool, AgentExecutor, LLMSingleActionAgent, AgentOutputParser
from langchain.prompts import StringPromptTemplate
from langchain import HuggingFaceHub, LLMChain
from typing import List, Union
from langchain.schema import AgentAction, AgentFinish, HumanMessage
import re
import mysql.connector

# Database configuration
db_config = {
    "host": "mariadb",
    "user": os.environ["MYSQL_USER"],
    "password": os.environ["MYSQL_PASSWORD"],
    "database": os.environ["MYSQL_DATABASE"],
}

# Define a function to update the database
def update_user_status(username: str, status: str):
    try:
        cnx = mysql.connector.connect(**db_config)
        cursor = cnx.cursor()

        update_query = """
            UPDATE Users 
            SET status = %s 
            WHERE username = %s
        """
        cursor.execute(update_query, (status, username))
        cnx.commit()

        return f"User {username} updated to status {status}"

    except mysql.connector.Error as err:
        return f"Error: {err}"

    finally:
        if cnx.is_connected():
            cursor.close()
            cnx.close()

# Define a tool for CRUD operations
tools = [
    Tool(
        name="CRUD Tool",
        func=update_user_status,
        description="Use this tool to execute CRUD operations on the database. The input should be in the format: <username> <status>. For example, 'john_doe active' to change user john_doe to active status.",
    )
]

# Set up the language model
llm = HuggingFaceHub(repo_id="google/flan-t5-xl", model_kwargs={"temperature":0.0}) 

# Define the prompt template
template = """You are a chatbot that can execute CRUD operations on a MariaDB database. 
The database has a table named 'Users' with columns: username, password, and status. 
Valid status values are 'active', 'pending', and 'closed'.

Available actions:
* You can use the CRUD Tool to change the status of a user.

Example:
User: Change user john_doe to status active.
You: CRUD Tool: john_doe active

Begin!

{input}"""
prompt = StringPromptTemplate(
    input_variables=["input"], template=template
)

# Define the output parser
class CustomOutputParser(AgentOutputParser):
    def parse(self, llm_output: str) -> Union[AgentAction, AgentFinish]:
        if "CRUD Tool:" in llm_output:
            # Extract username and status from the LLM output
            match = re.search(r"CRUD Tool: (\w+) (active|pending|closed)", llm_output)
            if match:
                username = match.group(1)
                status = match.group(2)
                return AgentAction(tool="CRUD Tool", tool_input=f"{username} {status}", log=llm_output)
            else:
                return AgentFinish(return_values={"output": "Invalid input format. Please use the format: <username> <status>"}, log=llm_output)
        else:
            return AgentFinish(return_values={"output": "No action needed."}, log=llm_output)

output_parser = CustomOutputParser()

# Set up the agent
llm_chain = LLMChain(llm=llm, prompt=prompt)
tool_names = [tool.name for tool in tools]
agent = LLMSingleActionAgent(
    llm_chain=llm_chain,
    output_parser=output_parser,
    stop=["\nObservation:"],
    allowed_tools=tool_names
)

# Create the agent executor
agent_executor = AgentExecutor.from_agent_and_tools(agent=agent, tools=tools, verbose=True)

# Run the chatbot
while True:
    user_input = input("User: ")
    response = agent_executor.run(user_input)
    print(response)