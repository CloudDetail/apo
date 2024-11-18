import { createStore } from 'redux';
import { persistStore } from 'redux-persist';
import rootReducer from './reducers/rootReducer';

const store = createStore(rootReducer);
const persistor = persistStore(store);

export { store, persistor };
