import { Container, Stack } from "@chakra-ui/react";
import Navbar from "./components/Navbar";
import TodoForm from "./components/TodoForm";
import TodoList from "./components/TodoList";


export const BASE_URL = "/api";

function App() {
	return (
		<Stack h='100vh' justifyContent={"center"} alignItems={"center"}>
			<Navbar />
			<Container>
				<TodoForm />
				<TodoList />
			</Container>
		</Stack>
	);
}

export default App;