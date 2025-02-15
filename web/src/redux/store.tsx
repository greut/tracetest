import {Action, configureStore, ThunkAction} from '@reduxjs/toolkit';
import TestAPI from 'redux/apis/TraceTest.api';
import TestDefinition from 'redux/slices/TestDefinition.slice';

export const store = configureStore({
  reducer: {
    [TestAPI.reducerPath]: TestAPI.reducer,
    testDefinition: TestDefinition,
  },
  middleware: getDefaultMiddleware => getDefaultMiddleware().concat(TestAPI.middleware),
});

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, Action<string>>;
