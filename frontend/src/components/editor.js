import React, { useState, useEffect, useRef } from 'react';
import './App.css'

const TextEditor = () => {
    const [text, setText] = useState('');
    const socketRef = useRef(null);

    useEffect(() => {
        // Connect to WebSocket endpoint
        socketRef.current = new WebSocket('ws://localhost:8000/ws');

        // Handle WebSocket events
        socketRef.current.onopen = () => {
            console.log('Connected to WebSocket');
        };

        socketRef.current.onmessage = (event) => {
            const receivedText = event.data;
            setText(receivedText);
        };

        socketRef.current.onclose = () => {
            console.log('Disconnected from WebSocket');
        };

        // Cleanup function
        return () => {
            // Close the WebSocket connection when the component is unmounted
            if (socketRef.current) {
                socketRef.current.close();
            }
        };
    }, []);

    const handleChange = (event) => {
        const newText = event.target.value;
        setText(newText);
        // Send the updated text to the WebSocket server
        if (socketRef.current) {
            socketRef.current.send(newText);
        }
    };

    return (
        <div className="document">
            {text ? (
                <textarea
                    className="document-textarea"
                    value={text}
                    onChange={handleChange}
                />
            ) : (
                <textarea
                    className="document-textarea document-textarea-placeholder"
                    placeholder="Start typing..."
                    onChange={handleChange}
                />
            )}
        </div>
    );
};

export default TextEditor