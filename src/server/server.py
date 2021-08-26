import asyncio
import websockets


async def server(websocket, path):
    # Get received data from websocket
    data = await websocket.recv()
    print(data)
    # Send response back to client to acknowledge receiving message
    await websocket.send("Thanks for your message: " + data)

if __name__ == "__main__":
    # Create websocket server
    start_server = websockets.serve(server, "localhost", 3000)
    # Start and run websocket server forever
    asyncio.get_event_loop().run_until_complete(start_server)
    asyncio.get_event_loop().run_forever()
