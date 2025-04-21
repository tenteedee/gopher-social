import { useNavigate, useParams } from 'react-router-dom';
import { API_URL } from './App';
import axios from 'axios';

export const ActivationPage = () => {
  const { token = '' } = useParams();
  const navigate = useNavigate();

  const handleActivate = async () => {
    const response = await axios.put(`${API_URL}/users/activate/${token}`);

    if (response.status === 204) {
      navigate('/');
    } else {
      // handle error
      alert('Failed to confirm token');
    }
  };

  return (
    <div>
      <h1>Activate</h1>
      <button onClick={handleActivate}>Click to activate</button>
    </div>
  );
};
