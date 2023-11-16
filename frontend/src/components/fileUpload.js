import React, { useState } from 'react';

const UploadComponent = () => {
    const [file, setFile] = useState(null);
    const [userID, setUserID] = useState('');
    const [responses, setResponses] = useState([]);
    const [componentKey, setComponentKey] = useState(0); // Unique key for the component

    const handleFileChange = (event) => {
        setFile(event.target.files[0]);
    };

    const handleUserIDChange = (event) => {
        setUserID(event.target.value);
    };

    const handleUpload = async () => {
        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('userID', userID);

            // Send file and userID to the upload API
            await fetch('http://localhost/api/upload', {
                method: 'POST',
                body: formData,
            });

            // Clear the file input and userID field
            setFile(null);
            setUserID('');

            // Increment the component key to force a reload
            setComponentKey((prevKey) => prevKey + 1);
        } catch (error) {
            console.error('Error occurred during file upload:', error);
        }

        // Set up WebSocket connection when uploading completes
        const ws = new WebSocket(`ws://localhost/notifications/ws/${userID}`);

        // Handle WebSocket message
        ws.onmessage = (event) => {
            const response = event.data;
            setResponses((prevResponses) => [...prevResponses, response]);
        };

        // Clean up WebSocket connection
        return () => {
            ws.close();
        };
    };

    return (
        <div key={componentKey} style={containerStyle}>
            <input
                type="text"
                placeholder="User ID"
                value={userID}
                onChange={handleUserIDChange}
                style={inputStyle}
            />
            <br />
            <h2>File Upload</h2>
            <input type="file" onChange={handleFileChange} style={inputStyle} />
            <br />
            <button onClick={handleUpload} style={buttonStyle}>
                Upload
            </button>
            <br />
            <h2>Responses:</h2>
            <div style={responseContainerStyle}>
                {responses.map((response, index) => (
                    <p key={index} style={responseStyle}>
                        {response}
                    </p>
                ))}
            </div>
        </div>
    );
};

// Styling
const containerStyle = {
    fontFamily: 'Arial, sans-serif',
    maxWidth: '400px',
    margin: '0 auto',
    padding: '20px',
};

const inputStyle = {
    padding: '5px',
    marginBottom: '10px',
};

const buttonStyle = {
    padding: '10px',
    backgroundColor: '#007bff',
    color: '#fff',
    border: 'none',
    borderRadius: '5px',
    cursor: 'pointer',
};

const responseContainerStyle = {
    border: '1px solid #ccc',
    padding: '10px',
};

const responseStyle = {
    margin: '5px 0',
};

export default UploadComponent;
