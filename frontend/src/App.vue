<template>
  <div id="app">
    <v-app>
      <v-app-bar absolute>
        <v-toolbar-title>confman-go</v-toolbar-title>
      </v-app-bar>

      <v-container>
        <v-row align="center" justify="center">
          <ServicePathConfigs />
        </v-row>
      </v-container>
    </v-app>
  </div>
</template>

<script>
import ServicePathConfigs from "./components/ServicePathConfigs.vue";

export default {
  name: "App",
  components: {
    ServicePathConfigs
  }
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
