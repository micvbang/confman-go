<template>
  <div>
    <h2>Service path configs</h2>
    <v-text-field id="service-path-input" v-model="servicePathFilter" label="Filter" />
    <v-card v-show="loading">
      <v-card-title>Fetching service path configs...</v-card-title>
      <v-card-text>
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
      </v-card-text>
    </v-card>

    <v-expansion-panels v-show="!loading" focusable>
      <v-expansion-panel v-for="servicePath in filteredServicePaths" :key="servicePath">
        <v-expansion-panel-header>{{servicePath}}</v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-dialog v-model="deleteDialog" width="500">
            <template v-slot:activator="{ on, attrs }">
              <v-col class="text-right">
                <v-btn small color="error" v-bind="attrs" v-on="on">Delete</v-btn>
              </v-col>
            </template>
            <v-card>
              <v-card-title class="headline grey lighten-2" primary-title>Delete</v-card-title>

              <v-card-text>
                You are about to delete all ({{ Object.keys(servicePathConfigs[servicePath]).length }}) keys configuration in
                <br />
                <b>{{ servicePath }}</b>
              </v-card-text>

              <v-divider></v-divider>

              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn
                  color="primary"
                  text
                  @click="servicePathConfigDeleteKeys(servicePath); deleteDialog = false"
                  :loading="deleteLoading"
                >Delete</v-btn>
                <v-btn color="secondary" text @click="deleteDialog = false">Cancel</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>

          <v-data-table
            :headers="[
                        {text: 'Key', align: 'start', value: 'key'},
                        {text: 'Value', align: 'start', value: 'value'}
                      ]"
            :items="servicePathConfigToDataTableItems(servicePathConfigs[servicePath])"
          ></v-data-table>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </div>
</template>

<script>
import { ConfmanClient } from "../clients/confman/client";

export default {
  name: "ServicePathConfigs",

  components: {},

  computed: {
    filteredServicePaths: function() {
      if (this.servicePathFilter.length === 0) {
        return this.servicePaths;
      }

      const lst = [];
      this.servicePaths.forEach(servicePath => {
        if (servicePath.includes(this.servicePathFilter)) {
          lst.push(servicePath);
        }
      });
      return lst;
    }
  },

  methods: {
    servicePathConfigToDataTableItems(servicePathConfig) {
      const lst = [];
      for (let key in servicePathConfig) {
        lst.push({ key: key, value: servicePathConfig[key] });
      }
      return lst;
    },

    async servicePathConfigDeleteKeys(servicePath) {
      this.deleteLoading = true;
      const keys = Object.keys(this.servicePathConfigs[servicePath]);
      await this.client.deleteServicePathKeys(servicePath, keys);
      this.deleteLoading = false;

      // TODO: this is expensive if there are many configs
      await this.loadServicePathConfigs();
    },

    async loadServicePathConfigs() {
      this.loading = true;
      this.servicePathConfigs = await this.client.getServicePathConfigs();
      this.servicePaths = Object.keys(this.servicePathConfigs).sort();
      this.loading = false;
    }
  },

  async created() {
    this.client = new ConfmanClient();
    await this.loadServicePathConfigs();
  },

  data: function() {
    return {
      client: null,
      servicePathFilter: "",
      servicePathConfigs: {},
      servicePaths: [],
      loading: true,
      deleteDialog: false,
      deleteLoading: false
    };
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="postcss" scoped>
$margin: 50px;

p {
  margin-bottom: 0;
  font-size: 1.2rem;
}

h2 {
  margin-top: 0;
  margin-bottom: 2rem;
  text-align: center;
}

.landingpage {
  background: #fefefe;
  color: #232323;
}

.section {
  display: flex;
  padding-top: $margin;
  padding-bottom: $margin;

  & h2 {
    font-size: 2.5rem;

    @media (--for-phone-only) {
      font-size: 2rem;
    }
  }
}

.section--columns {
  flex-direction: column;

  .section__text--big {
    padding-right: 0;
    margin-bottom: $margin;
  }
}

.section--inversed-colors {
  display: flex;
  width: 100%;
  justify-content: center;
  background: #333;
  color: var(--text);

  @media (--for-phone-only) {
    margin: 0;
  }
}

.section__center {
  width: 420px;
  align-self: center;
  text-align: center;

  @media (--for-phone-only) {
    width: 100%;
  }
}

.section__text {
  display: flex;
  flex-direction: column;
  align-content: center;
  justify-content: center;
}

.section__text--big {
  padding-right: $margin;
  text-align: center;

  @media (--for-tablet-portrait-down) {
    padding-right: 0;
    margin-bottom: $margin;
  }
}

.section__image {
  display: flex;
  flex-direction: column;
  align-content: center;
  justify-content: center;

  @media (--for-phone-only) {
    width: 100%;
  }
}

.section__image__select {
  display: flex;
  flex-wrap: wrap;
  margin-top: 20px;
  margin-right: -10px;
  margin-left: -10px;
}

.section__image__select__option {
  flex-grow: 1;
  padding: 10px;
  margin: 0 10px;
  background: #444;
  color: #bfbfbf;
  cursor: pointer;
  text-align: center;

  &.selected,
  &:hover {
    background: #2a2a2a;
    color: var(--text);
  }

  @media (--for-phone-only) {
    width: 100%;
    margin-bottom: 10px;
  }
}

.landingpage__video-player {
  width: 100%;
  color: var(--text);

  @media (--for-tablet-portrait-down) {
    width: unset;
    height: unset;
  }
}

.landingpage-content {
  display: flex;
  overflow: hidden;
  width: 100%;
  flex-direction: column;
  align-items: center;
}

.hero__container {
  position: relative;
  z-index: 1;
  display: flex;
  overflow: hidden;
  width: 100%;
  background: var(--background);
  color: var(--text);

  & .section {
    flex-wrap: wrap;
  }
}

.hero__subtitle {
  p {
    font-size: 1.5rem;
    line-height: 1.5;
  }
}

.hero__text-container {
  width: 100%;
  margin-bottom: 2rem;

  @media (--for-big-desktop-up) {
    flex: 2;
  }
}

.hero__video {
  width: 100%;
  flex: unset;
  margin-left: 0;

  @media (--for-big-desktop-up) {
    flex: 3;
    margin-left: 1rem;
  }
}

.hero__small-text {
  color: #bbb;
  font-size: 0.8rem;
}

.hero--image {
  position: absolute;
  z-index: -1;
  top: -56px;

  & img {
    height: 800px;
    transform: rotate(-45deg);
  }
}

.hero__title {
  & h1 {
    margin: 0;
    font-size: 3rem;

    @media (--for-phone-only) {
      font-size: 3rem;
    }
  }
}

.hero--overlay {
  position: absolute;
  z-index: -1;
  width: 100%;
  height: 100%;
  background: rgba(255, 255, 255, 0.2);
}

.landingpage__row {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  justify-content: space-around;
}

.landingpage__focus {
  width: 33%;
  box-sizing: border-box;
  padding: 1rem;
  line-height: 1.5;
  text-align: center;

  & h3 {
    margin-top: 5px;
  }

  @media (--for-tablet-portrait-down) {
    width: 100%;
  }
}

.landingpage__focus__number {
  width: 2rem;
  height: 2rem;
  padding: 1rem;
  margin: 0 auto;
  margin-bottom: 1rem;
  background: #e6eaec;
  border-radius: 50%;
  color: var(--noesisBlue);
  font-size: 2rem;
  font-weight: bold;
  line-height: 2rem;
  user-select: none;
}

.landingpage__display {
  display: flex;
  flex: 1;
  flex-direction: column;
}

.product-display {
  margin-top: 2rem;
}

.hero__cta {
  margin: 1rem 0;

  & a {
    display: inline-block;
    padding: 1rem;
    background: #fa8e06; /*var(--noesisBlue);*/
    border-radius: 5px;
    color: #1b1b1b;
    cursor: pointer;
    font-size: 1.3rem;
    font-weight: bold;
    text-decoration: none;

    &:hover,
    &.selected {
      transform: translateY(-1px);
    }

    &:hover,
    &:focus {
      box-shadow: 0 7px 14px 0 rgba(50, 50, 93, 0.1),
        0 3px 6px 0 rgba(0, 0, 0, 0.08);
    }

    &:active {
      transform: translateY(1px);
    }
  }
}

.landingpage__testimonial {
  width: 33%;
  box-sizing: border-box;
  flex-shrink: 0;
  padding: 1rem;
  text-align: center;

  @media (--for-tablet-landscape-down) {
    width: 50%;
  }

  @media (--for-tablet-portrait-down) {
    width: 100%;
  }
}

.landingpage__testimonial__quote {
  font-style: italic;
  line-height: 1.5;

  & i {
    font-size: 3rem;
  }
}

.landingpage__testimonial__source {
  display: flex;
  align-items: center;
  margin-top: 1rem;
  font-size: 0.9rem;
  text-align: right;
}

.landingpage__testimonial__source__image {
  margin-right: 1rem;
  margin-left: auto;

  & img {
    width: 3rem;

    &.circle-image {
      border-radius: 50%;
    }
  }
}

.landingpage__testimonial__source__name {
  font-weight: bold;
  text-align: right;
}

.landingpage__testimonial__navigate {
  position: absolute;
  z-index: 2;
  top: 0;
  left: -48px;
  display: flex;
  width: 48px;
  height: 100%;
  align-items: center;
  cursor: pointer;
  user-select: none;

  & i {
    font-size: 3rem;
  }

  &:last-child {
    right: -48px;
    left: unset;
  }

  @media (--for-phone-only) {
    left: 0;

    &:last-child {
      right: 0;
      left: unset;
    }
  }
}

.landingpage__testimonial__list {
  & span {
    display: flex;
    width: 100%;
  }

  @media (--for-phone-only) {
    padding-right: 2rem;
    padding-left: 2rem;
  }
}

.landingpage__testimonial {
  display: inline-block;
  transition: all 0.5s;

  &:last-child.testimonial-transition-enter {
    transform: translateX(100%);
  }

  &:first-child.testimonial-transition-leave-to {
    transform: translateX(-100%);
  }
}
.testimonial-transition-enter {
  opacity: 0;
  transform: translateX(-100%);
}
.testimonial-transition-leave-to {
  opacity: 0;
  transform: translateX(300%);
}
.testimonial-transition-leave-active {
  position: absolute;
}

.product-tour__row {
  display: flex;
  align-items: center;
  margin-bottom: 2rem;

  @media (--for-phone-only) {
    flex-direction: column;
  }
}
.product-tour__row--reverse {
  flex-direction: row-reverse;

  @media (--for-phone-only) {
    flex-direction: column;
  }
}

.product-tour__image {
  flex: 1;

  & img {
    width: 100%;
    border-radius: 100% 100% 100% 100%;
    box-shadow: 0 7px 14px 0 rgba(50, 50, 93, 0.2),
      0 3px 6px 0 rgba(0, 0, 0, 0.18);
  }
}

.product-tour__text {
  box-sizing: border-box;
  flex: 1;
  padding: 2rem;

  & h3 {
    margin: 0;
    margin-bottom: 1rem;
    font-size: 2rem;
  }

  & p {
    margin: 0;
    font-size: 1.1rem;
    line-height: 1.3;
  }

  &::before {
    display: block;
    width: 150px;
    height: 10px;
    margin-bottom: 1rem;
    background: var(--noesisBlue);
    content: "";
  }
}
</style>
