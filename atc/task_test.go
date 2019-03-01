package atc_test

import (
	. "github.com/concourse/concourse/atc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskConfig", func() {
	Describe("validating", func() {
		var (
			invalidConfig TaskConfig
			validConfig   TaskConfig
		)

		BeforeEach(func() {
			validConfig = TaskConfig{
				Platform: "linux",
				Run: TaskRunConfig{
					Path: "reboot",
				},
			}

			invalidConfig = validConfig
		})

		Describe("decode task yaml", func() {
			Context("given a valid task config", func() {
				It("works", func() {
					data := []byte(`
platform: beos

inputs: []

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(task.Platform).To(Equal("beos"))
					Expect(task.Run.Path).To(Equal("a/file"))
				})

				It("converts yaml booleans to strings in params", func() {
					data := []byte(`
platform: beos

params:
  testParam: true

run: {path: a/file}
`)
					config, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(config.Params["testParam"]).To(Equal("true"))
				})

				It("converts yaml ints to the correct string in params", func() {
					data := []byte(`
platform: beos

params:
  testParam: 1059262

run: {path: a/file}
`)
					config, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(config.Params["testParam"]).To(Equal("1059262"))
				})

				It("converts yaml floats to the correct string in params", func() {
					data := []byte(`
platform: beos

params:
  testParam: 1059262.123123123

run: {path: a/file}
`)
					config, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(config.Params["testParam"]).To(Equal("1059262.123123123"))
				})

				It("converts maps to json in params", func() {
					data := []byte(`
platform: beos

params:
  testParam:
    foo: bar

run: {path: a/file}
`)
					config, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(config.Params["testParam"]).To(Equal(`{"foo":"bar"}`))
				})
			})

			Context("given a valid task config with numeric params", func() {
				It("works", func() {
					data := []byte(`
platform: beos

params:
  FOO: 1

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(task.Platform).To(Equal("beos"))
					Expect(task.Params).To(Equal(map[string]string{"FOO": "1"}))
				})
			})

			Context("given a valid task config with extra keys", func() {
				It("returns an error", func() {
					data := []byte(`
platform: beos

intputs: []

run: {path: a/file}
`)
					_, err := NewTaskConfig(data)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("given an invalid task config", func() {
				It("errors on validation", func() {
					data := []byte(`
platform: beos

inputs: ['a/b/c']
outputs: ['a/b/c']

run: {path: a/file}
`)
					_, err := NewTaskConfig(data)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when platform is missing", func() {
			BeforeEach(func() {
				invalidConfig.Platform = ""
			})

			It("returns an error", func() {
				Expect(invalidConfig.Validate()).To(MatchError(ContainSubstring("  missing 'platform'")))
			})
		})

		Context("when container limits are specified", func() {
			Context("when memory and cpu limits are correctly specified", func() {
				It("successfully parses the limits with memory units", func() {
					data := []byte(`
platform: beos
container_limits: { cpu: 1024, memory: 1KB }

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					cpu := uint64(1024)
					memory := uint64(1024)
					Expect(task.Limits).To(Equal(ContainerLimits{
						CPU:    &cpu,
						Memory: &memory,
					}))
				})

				It("successfully parses the limits without memory units", func() {
					data := []byte(`
platform: beos
container_limits: { cpu: 1024, memory: 209715200 }

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					cpu := uint64(1024)
					memory := uint64(209715200)
					Expect(task.Limits).To(Equal(ContainerLimits{
						CPU:    &cpu,
						Memory: &memory,
					}))
				})
			})

			Context("when either one of memory or cpu is correctly specified", func() {
				It("parses the provided memory limit without any errors", func() {
					data := []byte(`
platform: beos
container_limits: { memory: 1KB }

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					memory := uint64(1024)
					Expect(task.Limits).To(Equal(ContainerLimits{
						Memory: &memory,
					}))
				})

				It("parses the provided cpu limit without any errors", func() {
					data := []byte(`
platform: beos
container_limits: { cpu: 355 }

run: {path: a/file}
`)
					task, err := NewTaskConfig(data)
					Expect(err).ToNot(HaveOccurred())
					cpu := uint64(355)
					Expect(task.Limits).To(Equal(ContainerLimits{
						CPU: &cpu,
					}))
				})
			})

			Context("when invalid memory limit value is provided", func() {
				It("throws an error and does not continue", func() {
					data := []byte(`
platform: beos
container_limits: { cpu: 1024, memory: abc1000kb  }

run: {path: a/file}
`)
					_, err := NewTaskConfig(data)
					Expect(err).To(MatchError(ContainSubstring("could not parse container memory limit")))
				})

			})

			Context("when invalid cpu limit value is provided", func() {
				It("throws an error and does not continue", func() {
					data := []byte(`
platform: beos
container_limits: { cpu: str1ng-cpu-l1mit, memory: 20MB}

run: {path: a/file}
`)
					_, err := NewTaskConfig(data)
					Expect(err).To(MatchError(ContainSubstring("cpu limit must be an integer")))
				})
			})
		})

		Context("when the task has inputs", func() {
			BeforeEach(func() {
				validConfig.Inputs = append(validConfig.Inputs, TaskInputConfig{Name: "concourse"})
			})

			It("is valid", func() {
				Expect(validConfig.Validate()).ToNot(HaveOccurred())
			})

			Context("when input.name is missing", func() {
				BeforeEach(func() {
					invalidConfig.Inputs = append(invalidConfig.Inputs, TaskInputConfig{Name: "concourse"}, TaskInputConfig{Name: ""})
				})

				It("returns an error", func() {
					Expect(invalidConfig.Validate()).To(MatchError(ContainSubstring("  input in position 1 is missing a name")))
				})
			})

			Context("when input.name is missing multiple times", func() {
				BeforeEach(func() {
					invalidConfig.Inputs = append(
						invalidConfig.Inputs,
						TaskInputConfig{Name: "concourse"},
						TaskInputConfig{Name: ""},
						TaskInputConfig{Name: ""},
					)
				})

				It("returns an error", func() {
					err := invalidConfig.Validate()

					Expect(err).To(MatchError(ContainSubstring("  input in position 1 is missing a name")))
					Expect(err).To(MatchError(ContainSubstring("  input in position 2 is missing a name")))
				})
			})
		})

		Context("when the task has outputs", func() {
			BeforeEach(func() {
				validConfig.Outputs = append(validConfig.Outputs, TaskOutputConfig{Name: "concourse"})
			})

			It("is valid", func() {
				Expect(validConfig.Validate()).ToNot(HaveOccurred())
			})

			Context("when output.name is missing", func() {
				BeforeEach(func() {
					invalidConfig.Outputs = append(invalidConfig.Outputs, TaskOutputConfig{Name: "concourse"}, TaskOutputConfig{Name: ""})
				})

				It("returns an error", func() {
					Expect(invalidConfig.Validate()).To(MatchError(ContainSubstring("  output in position 1 is missing a name")))
				})
			})

			Context("when output.name is missing multiple times", func() {
				BeforeEach(func() {
					invalidConfig.Outputs = append(
						invalidConfig.Outputs,
						TaskOutputConfig{Name: "concourse"},
						TaskOutputConfig{Name: ""},
						TaskOutputConfig{Name: ""},
					)
				})

				It("returns an error", func() {
					err := invalidConfig.Validate()

					Expect(err).To(MatchError(ContainSubstring("  output in position 1 is missing a name")))
					Expect(err).To(MatchError(ContainSubstring("  output in position 2 is missing a name")))
				})
			})
		})

		Context("when run is missing", func() {
			BeforeEach(func() {
				invalidConfig.Run.Path = ""
			})

			It("returns an error", func() {
				Expect(invalidConfig.Validate()).To(MatchError(ContainSubstring("  missing path to executable to run")))
			})
		})

	})

	Describe("merging", func() {
		It("merges params while preserving other properties", func() {
			Expect(TaskConfig{
				RootfsURI: "some-image",
				Params: map[string]string{
					"FOO": "1",
					"BAR": "2",
				},
			}.Merge(TaskConfig{
				Params: map[string]string{
					"FOO": "3",
				},
			})).To(

				Equal(TaskConfig{
					RootfsURI: "some-image",
					Params: map[string]string{
						"FOO": "3",
						"BAR": "2",
					},
				}))
		})

		It("emits a warning if params key in pipeline.yml is not defined in task.yml", func() {
			config, warnings, err := TaskConfig{
				RootfsURI: "some-image",
				Params: map[string]string{
					"FOO": "1",
					"BAR": "2",
				},
			}.Merge(TaskConfig{
				Params: map[string]string{
					"FOO": "3",
					"BAZ": "4",
				},
			})

			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("BAZ was defined in pipeline but missing from task file"))
			Expect(config.Params).To(Equal(map[string]string{
				"FOO": "3",
				"BAR": "2",
				"BAZ": "4",
			}))
			Expect(err).To(BeNil())
		})

		It("merges into nil params without panicking", func() {
			merged, warnings, err := TaskConfig{
				Params: map[string]string{
					"FOO": "3",
				},
			}.Merge(TaskConfig{
				RootfsURI: "some-image",
			})
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
			Expect(merged).To(Equal(TaskConfig{
				RootfsURI: "some-image",
				Params: map[string]string{
					"FOO": "3",
				},
			}))

			merged, warnings, err = TaskConfig{
				RootfsURI: "some-image",
			}.Merge(TaskConfig{
				Params: map[string]string{
					"FOO": "3",
				},
			})
			Expect(err).To(BeNil())
			Expect(merged).To(Equal(TaskConfig{
				RootfsURI: "some-image",
				Params: map[string]string{
					"FOO": "3",
				},
			}))
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("FOO was defined in pipeline but missing from task file"))
		})

		It("overrides the platform", func() {
			Expect(TaskConfig{
				Platform: "platform-a",
			}.Merge(TaskConfig{
				Platform: "platform-b",
			})).To(

				Equal(TaskConfig{
					Platform: "platform-b",
				}))

		})

		It("overrides the image", func() {
			Expect(TaskConfig{
				RootfsURI: "some-image",
			}.Merge(TaskConfig{
				RootfsURI: "better-image",
			})).To(

				Equal(TaskConfig{
					RootfsURI: "better-image",
				}))

		})

		It("overrides the run config", func() {
			Expect(TaskConfig{
				Run: TaskRunConfig{
					Path: "some-path",
					Args: []string{"arg1", "arg2"},
				},
			}.Merge(TaskConfig{
				RootfsURI: "some-image",
				Run: TaskRunConfig{
					Path: "better-path",
					Args: []string{"better-arg1", "better-arg2"},
				},
			})).To(

				Equal(TaskConfig{
					RootfsURI: "some-image",
					Run: TaskRunConfig{
						Path: "better-path",
						Args: []string{"better-arg1", "better-arg2"},
					},
				}))

		})

		It("overrides the run config even with no args", func() {
			Expect(TaskConfig{
				Run: TaskRunConfig{
					Path: "some-path",
					Args: []string{"arg1", "arg2"},
				},
			}.Merge(TaskConfig{
				RootfsURI: "some-image",
				Run: TaskRunConfig{
					Path: "better-path",
				},
			})).To(

				Equal(TaskConfig{
					RootfsURI: "some-image",
					Run: TaskRunConfig{
						Path: "better-path",
					},
				}))

		})

		It("overrides input configuration", func() {
			Expect(TaskConfig{
				Inputs: []TaskInputConfig{
					{Name: "some-input", Path: "some-destination"},
				},
			}.Merge(TaskConfig{
				Inputs: []TaskInputConfig{
					{Name: "another-input", Path: "another-destination"},
				},
			})).To(

				Equal(TaskConfig{
					Inputs: []TaskInputConfig{
						{Name: "another-input", Path: "another-destination"},
					},
				}))

		})
	})
})