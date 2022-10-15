import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

import Signin from '../screens/Login';
import Signup from '../screens/Registration';
import Splash from '../screens/Splash';
import CreateRoom from '../screens/CreateRoom';
import RoomsPage from '../screens/RoomsPage';
import BottomTab from './BottomTab';
import JoinRoom from '../screens/JoinRoom';
import GetSupermarkets from '../screens/GetSupermarkets';
import SplitBillMembers from '../screens/SplitBillMembers';
import SplitBill from '../screens/SplitBill';

const Navigation = () => {
  const Stack = createNativeStackNavigator();

  return (
    <NavigationContainer>
      <Stack.Navigator
        initialRouteName="Splash"
        screenOptions={{ headerShown: false }}>
        <Stack.Screen name="Splash" component={Splash} />
        <Stack.Screen name="Signin" component={Signin} />
        <Stack.Screen name="Signup" component={Signup} />
        <Stack.Screen name="HomeTab" component={BottomTab} />
        <Stack.Screen name="CreateRoom" component={CreateRoom} />
        <Stack.Screen name="RoomsPage" component={RoomsPage} />
        <Stack.Screen name="JoinRoom" component={JoinRoom} />
        <Stack.Screen name="SplitBill" component={SplitBill} />
        <Stack.Screen name="SplitBillMembers" component={SplitBillMembers} />
        <Stack.Screen name="GetSupermarkets" component={GetSupermarkets} />
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default Navigation;
