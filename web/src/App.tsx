import React, {useEffect, useState} from 'react';
import { Admin } from './admin/Admin';
import './App.css';

function App() {
   return <Admin server={"/api"}/>
}


export default App;
