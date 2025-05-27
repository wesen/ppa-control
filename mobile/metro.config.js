const { getDefaultConfig } = require('expo/metro-config');

/** @type {import('expo/metro-config').MetroConfig} */
const config = getDefaultConfig(__dirname);

// Add support for resolving react-native-udp
config.resolver.alias = {
  ...config.resolver.alias,
  'react-native-udp': require.resolve('react-native-udp'),
};

module.exports = config;